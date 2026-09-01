package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory-chain/internal/genai"
)

type stubBlockchainClient struct {
	issuedDeptID    string
	issuedName      string
	issuedCategory  string
	issuedQty       int
	issuedThreshold int
}

func (s *stubBlockchainClient) IssueAsset(deptID, name, category string, qty, threshold int) (string, string, time.Time, error) {
	s.issuedDeptID = deptID
	s.issuedName = name
	s.issuedCategory = category
	s.issuedQty = qty
	s.issuedThreshold = threshold
	return "tx-1", "asset-1", time.Now().UTC(), nil
}

func (s *stubBlockchainClient) ConsumeStock(deptID, assetID string, qty int, purpose string) (string, int, bool, error) {
	return "", 0, false, nil
}

func (s *stubBlockchainClient) TransferAsset(fromDept, toDept, assetID string, qty int) (string, int, int, error) {
	return "", 0, 0, nil
}

func (s *stubBlockchainClient) GetAssetHistory(assetID string) ([]ClientHistoryEntry, error) {
	return nil, nil
}

func (s *stubBlockchainClient) ReadAsset(assetID string) (*ClientAsset, error) {
	return nil, nil
}

func (s *stubBlockchainClient) QueryAssetsByDept(deptID string) ([]*ClientAsset, error) {
	return nil, nil
}

func (s *stubBlockchainClient) RequestReplenishment(assetID string, qty int, urgency string) (string, error) {
	return "", nil
}

func (s *stubBlockchainClient) GetLedgerBlocks() ([]interface{}, error) {
	return nil, nil
}

func (s *stubBlockchainClient) VerifyLedger() (bool, *int, error) {
	return true, nil, nil
}

func (s *stubBlockchainClient) ClassifyPriority(assetID string, scores genai.PriorityScores) (string, string, float64, error) {
	return "", "", 0, nil
}

func (s *stubBlockchainClient) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	return "", nil
}

func (s *stubBlockchainClient) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	return "", nil
}

func (s *stubBlockchainClient) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	return "", nil
}

func (s *stubBlockchainClient) RetireAsset(assetID, reason string) (string, error) {
	return "", nil
}

func (s *stubBlockchainClient) AssistantQuery(userQuery string) (string, error) {
	return "", nil
}

func (s *stubBlockchainClient) QueryAssetsByPriority(tier string) ([]*ClientAsset, error) {
	return nil, nil
}

func (s *stubBlockchainClient) ComputeUtilization(assetID string) (float64, error) {
	return 0, nil
}

func TestIssueAssetAcceptsCamelCasePayload(t *testing.T) {
	client := &stubBlockchainClient{}
	h := NewHandlers(client)

	body := bytes.NewBufferString(`{"deptId":"Lab","name":"Monitor","category":"Medical","qty":5,"threshold":2}`)
	req := httptest.NewRequest(http.MethodPost, "/api/assets/issue", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.IssueAsset(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if client.issuedDeptID != "Lab" {
		t.Fatalf("expected deptID to be parsed, got %q", client.issuedDeptID)
	}
	if client.issuedQty != 5 {
		t.Fatalf("expected qty to be parsed, got %d", client.issuedQty)
	}
	if client.issuedThreshold != 2 {
		t.Fatalf("expected threshold to be parsed, got %d", client.issuedThreshold)
	}
}
