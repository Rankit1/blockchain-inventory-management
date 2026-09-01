package fabricclient

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type FabricClient struct {
	gateway    *client.Gateway
	network    *client.Network
	contract   *client.Contract
	grpcConn   *grpc.ClientConn
}

type Asset struct {
	AssetID          string    `json:"assetId"`
	DeptID           string    `json:"deptId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int       `json:"qty"`
	Threshold        int       `json:"threshold"`
	PriorityTier     string    `json:"priorityTier,omitempty"`
	CriticalityScore float64   `json:"criticalityScore,omitempty"`
	LifecycleState   string    `json:"lifecycleState,omitempty"`
	LastAuditDate    string    `json:"lastAuditDate,omitempty"`
	UtilizationRate  float64   `json:"utilizationRate,omitempty"`
	WarrantyExpiry   string    `json:"warrantyExpiry,omitempty"`
	AMCExpiry        string    `json:"amcExpiry,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type HistoryEntry struct {
	TxID      string    `json:"txId"`
	Value     *Asset    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

func Connect(cryptoDir string, peerAddress string, peerName string) (*FabricClient, error) {
	// 1. Locate Admin credentials
	userDir := filepath.Join(cryptoDir, "peerOrganizations", "inventory.com", "users", "Admin@inventory.com")
	certPath := filepath.Join(userDir, "msp", "signcerts", "Admin@inventory.com-cert.pem")
	keyDir := filepath.Join(userDir, "msp", "keystore")

	// Read client cert
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read client cert: %w", err)
	}
	cert, err := identity.CertificateFromPEM(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate PEM: %w", err)
	}
	id, err := identity.NewX509Identity("InventoryOrgMSP", cert)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 identity: %w", err)
	}

	// Read client key dynamically (find the key file in the keystore directory)
	files, err := ioutil.ReadDir(keyDir)
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("failed to read keystore directory: %w", err)
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	privateKey, err := identity.PrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key PEM: %w", err)
	}
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// 2. Load TLS CA Certificate
	caCertPath := filepath.Join(cryptoDir, "peerOrganizations", "inventory.com", "peers", peerName, "tls", "ca.crt")
	caCertBytes, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read peer CA TLS cert: %w", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertBytes) {
		return nil, fmt.Errorf("failed to append peer CA TLS cert")
	}

	// 3. Dial gRPC connection to peer
	creds := credentials.NewClientTLSFromCert(certPool, peerName)
	grpcConn, err := grpc.Dial(peerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer gRPC address: %w", err)
	}

	// 4. Connect Gateway client
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("failed to connect gateway: %w", err)
	}

	net := gw.GetNetwork("inventorychannel")
	contract := net.GetContract("assetcc")

	return &FabricClient{
		gateway:  gw,
		network:  net,
		contract: contract,
		grpcConn: grpcConn,
	}, nil
}

func (f *FabricClient) Close() {
	if f.gateway != nil {
		f.gateway.Close()
	}
	if f.grpcConn != nil {
		f.grpcConn.Close()
	}
}

// Helper method to submit transaction proposals with automatic retry on MVCC / commit conflicts
func (f *FabricClient) submitTransaction(txName string, args ...string) (string, []byte, error) {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(150*attempt) * time.Millisecond)
		}
		proposal, err := f.contract.NewProposal(txName, client.WithArguments(args...))
		if err != nil {
			lastErr = err
			continue
		}
		endorsed, err := proposal.Endorse()
		if err != nil {
			lastErr = err
			continue
		}
		submitted, err := endorsed.Submit()
		if err != nil {
			lastErr = err
			continue
		}
		status, err := submitted.Status()
		if err != nil || !status.Successful {
			lastErr = fmt.Errorf("transaction %s commit failed (code: %v, err: %v)", txName, status.Code, err)
			continue
		}
		return submitted.TransactionID(), endorsed.Result(), nil
	}
	return "", nil, lastErr
}

// REST SDK wrappers
func (f *FabricClient) IssueAsset(deptID, name, category string, qty, threshold int) (string, string, time.Time, error) {
	// Generate UUID asset ID
	assetID := fmt.Sprintf("asset-%d", time.Now().UnixNano())
	
	txID, _, err := f.submitTransaction("IssueAsset", assetID, deptID, name, category, fmt.Sprintf("%d", qty), fmt.Sprintf("%d", threshold))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return txID, assetID, time.Now().UTC(), nil
}

func (f *FabricClient) ConsumeStock(deptID, assetID string, qty int, purpose string) (string, int, bool, error) {
	var res struct {
		NewQty             int  `json:"newQty"`
		ReplenishTriggered bool `json:"replenishTriggered"`
	}

	txID, resultBytes, err := f.submitTransaction("ConsumeStock", assetID, fmt.Sprintf("%d", qty), purpose)
	if err != nil {
		return "", 0, false, err
	}

	// Parse the response from the submitted tx execution
	if err := json.Unmarshal(resultBytes, &res); err != nil {
		return "", 0, false, fmt.Errorf("failed to unmarshal consume result: %w", err)
	}

	return txID, res.NewQty, res.ReplenishTriggered, nil
}

func (f *FabricClient) TransferAsset(fromDept, toDept, assetID string, qty int) (string, int, int, error) {
	txID, _, err := f.submitTransaction("TransferAsset", fromDept, toDept, assetID, fmt.Sprintf("%d", qty))
	if err != nil {
		return "", 0, 0, err
	}

	// Retrieve quantities to return in response
	srcAsset, err := f.ReadAsset(assetID)
	if err != nil {
		return "", 0, 0, err
	}
	
	// Query toDept copy
	toAssets, err := f.QueryAssetsByDept(toDept)
	var toQty int
	if err == nil {
		for _, a := range toAssets {
			if a.Name == srcAsset.Name && a.Category == srcAsset.Category {
				toQty = a.Qty
				break
			}
		}
	}

	return txID, srcAsset.Qty, toQty, nil
}

func (f *FabricClient) ReadAsset(assetID string) (*Asset, error) {
	resultBytes, err := f.contract.EvaluateTransaction("ReadAsset", assetID)
	if err != nil {
		return nil, err
	}

	var asset Asset
	if err := json.Unmarshal(resultBytes, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (f *FabricClient) QueryAssetsByDept(deptID string) ([]*Asset, error) {
	resultBytes, err := f.contract.EvaluateTransaction("QueryAssetsByDept", deptID)
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if len(resultBytes) == 0 {
		return assets, nil
	}
	if err := json.Unmarshal(resultBytes, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

// QueryAssetsByPriority returns all assets matching a priority tier via CouchDB.
func (f *FabricClient) QueryAssetsByPriority(tier string) ([]*Asset, error) {
	resultBytes, err := f.contract.EvaluateTransaction("QueryAssetsByPriority", tier)
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if len(resultBytes) == 0 {
		return assets, nil
	}
	if err := json.Unmarshal(resultBytes, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func (f *FabricClient) GetAssetHistory(assetID string) ([]HistoryEntry, error) {
	resultBytes, err := f.contract.EvaluateTransaction("GetAssetHistory", assetID)
	if err != nil {
		return nil, err
	}

	var history []HistoryEntry
	if len(resultBytes) == 0 {
		return history, nil
	}
	if err := json.Unmarshal(resultBytes, &history); err != nil {
		return nil, err
	}
	return history, nil
}

func (f *FabricClient) RequestReplenishment(assetID string, qty int, urgency string) (string, error) {
	txID, _, err := f.submitTransaction("RequestReplenishment", assetID, fmt.Sprintf("%d", qty), urgency)
	return txID, err
}

// ClassifyPriority submits a priority classification transaction.
func (f *FabricClient) ClassifyPriority(assetID string, businessCriticality, replacementCost, replacementLeadTime, safetyComplianceImpact, redundancyAvailability int) (string, string, float64, error) {
	txID, resultBytes, err := f.submitTransaction("ClassifyPriority",
		assetID,
		fmt.Sprintf("%d", businessCriticality),
		fmt.Sprintf("%d", replacementCost),
		fmt.Sprintf("%d", replacementLeadTime),
		fmt.Sprintf("%d", safetyComplianceImpact),
		fmt.Sprintf("%d", redundancyAvailability),
	)
	if err != nil {
		return "", "", 0, err
	}

	var res struct {
		PriorityTier    string  `json:"priorityTier"`
		CriticalityScore float64 `json:"criticalityScore"`
	}
	if err := json.Unmarshal(resultBytes, &res); err != nil {
		return txID, "", 0, nil
	}
	return txID, res.PriorityTier, res.CriticalityScore, nil
}

// UpdatePriorityTier submits a manual priority override transaction.
func (f *FabricClient) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	txID, _, err := f.submitTransaction("UpdatePriorityTier", assetID, tier, reason)
	return txID, err
}

// ScheduleAudit submits an audit scheduling transaction.
func (f *FabricClient) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	txID, _, err := f.submitTransaction("ScheduleAudit", assetID, auditDate, scope)
	return txID, err
}

// RecordAuditResult submits an audit result transaction.
func (f *FabricClient) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	txID, _, err := f.submitTransaction("RecordAuditResult", assetID, auditDate, result, notes)
	return txID, err
}

// RetireAsset submits an asset retirement transaction.
func (f *FabricClient) RetireAsset(assetID, reason string) (string, error) {
	txID, _, err := f.submitTransaction("RetireAsset", assetID, reason)
	return txID, err
}
