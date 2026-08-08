package main

import (
	"encoding/json"
	"testing"
	"time"

	tproto "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type MockStub struct {
	shim.ChaincodeStubInterface
	state  map[string][]byte
	txID   string
	events map[string][]byte
}

func (m *MockStub) GetState(key string) ([]byte, error) {
	return m.state[key], nil
}

func (m *MockStub) PutState(key string, value []byte) error {
	m.state[key] = value
	return nil
}

func (m *MockStub) GetTxTimestamp() (*tproto.Timestamp, error) {
	return &tproto.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}, nil
}

func (m *MockStub) GetTxID() string {
	if m.txID == "" {
		return "test-tx-12345678"
	}
	return m.txID
}

func (m *MockStub) SetEvent(name string, payload []byte) error {
	m.events[name] = payload
	return nil
}

type MockContext struct {
	contractapi.TransactionContextInterface
	stub *MockStub
}

func (m *MockContext) GetStub() shim.ChaincodeStubInterface {
	return m.stub
}

func TestIssueAsset(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{
		state:  make(map[string][]byte),
		events: make(map[string][]byte),
	}
	ctx := &MockContext{stub: stub}

	err := contract.IssueAsset(ctx, "asset1", "Lab", "Beaker", "Glassware", 10, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, _ := stub.GetState("asset1")
	if val == nil {
		t.Fatal("expected asset1 to be saved in world state")
	}

	var asset Asset
	json.Unmarshal(val, &asset)
	if asset.Name != "Beaker" || asset.Qty != 10 {
		t.Errorf("incorrect asset fields: %+v", asset)
	}
}

func TestConsumeStock(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{
		state:  make(map[string][]byte),
		events: make(map[string][]byte),
	}
	ctx := &MockContext{stub: stub}

	// Preset state
	asset := Asset{
		DocType:   "asset",
		AssetID:   "asset1",
		DeptID:    "Lab",
		Name:      "Beaker",
		Category:  "Glassware",
		Qty:       10,
		Threshold: 3,
	}
	assetBytes, _ := json.Marshal(asset)
	stub.PutState("asset1", assetBytes)

	// Consume without trigger
	resp, err := contract.ConsumeStock(ctx, "asset1", 5, "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var res struct {
		NewQty             int  `json:"newQty"`
		ReplenishTriggered bool `json:"replenishTriggered"`
	}
	json.Unmarshal([]byte(resp), &res)
	if res.NewQty != 5 || res.ReplenishTriggered {
		t.Errorf("unexpected consume response: %+v", res)
	}

	// Consume and trigger replenishment
	resp2, _ := contract.ConsumeStock(ctx, "asset1", 3, "test2")
	json.Unmarshal([]byte(resp2), &res)
	if res.NewQty != 2 || !res.ReplenishTriggered {
		t.Errorf("expected replenishment trigger: %+v", res)
	}
	if stub.events["REPLENISH_REQUEST"] == nil {
		t.Error("expected REPLENISH_REQUEST event to be set")
	}
}
