package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	fmt.Println("====================================================")
	fmt.Println("   BLOCKCHAIN INVENTORY MANAGEMENT CLI CLIENT       ")
	fmt.Println("====================================================")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println()
		fmt.Println("----------------- MAIN MENU -----------------")
		fmt.Println("1. Issue New Asset")
		fmt.Println("2. Consume Stock")
		fmt.Println("3. Transfer Asset")
		fmt.Println("4. List All Assets")
		fmt.Println("5. View Asset Transaction History")
		fmt.Println("6. Verify Ledger Integrity")
		fmt.Println("7. Dump All Ledger Blocks")
		fmt.Println("8. Classify Asset Priority")
		fmt.Println("9. Update Priority Tier (Override)")
		fmt.Println("10. Schedule Audit")
		fmt.Println("11. Record Audit Result")
		fmt.Println("12. Retire Asset")
		fmt.Println("13. Utilization Report")
		fmt.Println("14. Compliance Report")
		fmt.Println("15. Exit")
		fmt.Println("---------------------------------------------")

		choice := readString(scanner, "Choose an option (1-15): ", true)

		switch choice {
		case "1":
			fmt.Println("\n--- Issue New Asset ---")
			dept := readString(scanner, "Department ID (e.g. Lab, IT, Store): ", true)
			name := readString(scanner, "Asset Name: ", true)
			cat := readString(scanner, "Category: ", true)
			qty := readInt(scanner, "Initial Quantity: ", 1)
			threshold := readInt(scanner, "Replenishment Threshold: ", 0)

			res, err := issueAsset(dept, name, cat, qty, threshold)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to issue asset: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Asset issued successfully!\n")
				fmt.Printf("  TxID:     %s\n", res.TxID)
				fmt.Printf("  AssetID:  %s\n", res.AssetID)
				fmt.Printf("  Time:     %s\n", res.Timestamp.Format(time.RFC3339))
			}

		case "2":
			fmt.Println("\n--- Consume Stock ---")
			dept := readString(scanner, "Department ID: ", true)
			assetID := readString(scanner, "Asset ID: ", true)
			qty := readInt(scanner, "Quantity to Consume: ", 1)
			purpose := readString(scanner, "Purpose: ", true)

			res, err := consumeStock(dept, assetID, qty, purpose)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to consume stock: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Stock consumed successfully!\n")
				fmt.Printf("  TxID:                %s\n", res.TxID)
				fmt.Printf("  New Quantity:        %d\n", res.NewQty)
				fmt.Printf("  Replenish Triggered: %t\n", res.ReplenishTriggered)
			}

		case "3":
			fmt.Println("\n--- Transfer Asset ---")
			fromDept := readString(scanner, "From Department ID: ", true)
			toDept := readString(scanner, "To Department ID: ", true)
			assetID := readString(scanner, "Asset ID: ", true)
			qty := readInt(scanner, "Quantity to Transfer: ", 1)

			res, err := transferAsset(fromDept, toDept, assetID, qty)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to transfer asset: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Asset transferred successfully!\n")
				fmt.Printf("  TxID:         %s\n", res.TxID)
				fmt.Printf("  FromDept Qty: %d\n", res.FromQty)
				fmt.Printf("  ToDept Qty:   %d\n", res.ToQty)
			}

		case "4":
			fmt.Println("\n--- Current World State (All Assets) ---")
			if err := listAssets(); err != nil {
				fmt.Printf("\n[ERROR] Failed to list assets: %v\n", err)
			}

		case "5":
			fmt.Println("\n--- View Asset Transaction History ---")
			assetID := readString(scanner, "Asset ID: ", true)

			if err := getHistory(assetID); err != nil {
				fmt.Printf("\n[ERROR] Failed to fetch history: %v\n", err)
			}

		case "6":
			fmt.Println("\n--- Verifying Ledger Chain Integrity ---")
			if err := verifyLedger(); err != nil {
				fmt.Printf("\n[ERROR] Failed to verify ledger: %v\n", err)
			}

		case "7":
			fmt.Println("\n--- Dumping All Ledger Blocks ---")
			if err := dumpBlocks(); err != nil {
				fmt.Printf("\n[ERROR] Failed to dump ledger blocks: %v\n", err)
			}

		case "8":
			fmt.Println("\n--- Classify Asset Priority ---")
			assetID := readString(scanner, "Asset ID: ", true)
			bc := readScore(scanner, "Business Criticality (1-5): ")
			rc := readScore(scanner, "Replacement Cost (1-5): ")
			lt := readScore(scanner, "Replacement Lead Time (1-5): ")
			sc := readScore(scanner, "Safety/Compliance Impact (1-5): ")
			ra := readScore(scanner, "Redundancy/Availability (1-5): ")

			res, err := classifyAsset(assetID, bc, rc, lt, sc, ra)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to classify asset: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Asset classified!\n")
				fmt.Printf("  TxID:          %s\n", res.TxID)
				fmt.Printf("  Priority Tier: %s\n", res.PriorityTier)
				fmt.Printf("  Criticality:   %.2f\n", res.CriticalityScore)
			}

		case "9":
			fmt.Println("\n--- Update Priority Tier (Manual Override) ---")
			assetID := readString(scanner, "Asset ID: ", true)
			tier := readString(scanner, "New Priority Tier (P1/P2/P3): ", true)
			reason := readString(scanner, "Justification: ", true)

			txID, err := updatePriority(assetID, tier, reason)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to update priority: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Priority updated! TxID: %s\n", txID)
			}

		case "10":
			fmt.Println("\n--- Schedule Audit ---")
			assetID := readString(scanner, "Asset ID: ", true)
			auditDate := readString(scanner, "Audit Date (YYYY-MM-DD): ", true)
			scope := readString(scanner, "Scope: ", false)

			txID, err := scheduleAudit(assetID, auditDate, scope)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to schedule audit: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Audit scheduled! TxID: %s\n", txID)
			}

		case "11":
			fmt.Println("\n--- Record Audit Result ---")
			assetID := readString(scanner, "Asset ID: ", true)
			auditDate := readString(scanner, "Audit Date (YYYY-MM-DD): ", true)
			result := readString(scanner, "Result (e.g. PASS/FAIL): ", true)
			notes := readString(scanner, "Notes: ", false)

			txID, err := recordAudit(assetID, auditDate, result, notes)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to record audit: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Audit result recorded! TxID: %s\n", txID)
			}

		case "12":
			fmt.Println("\n--- Retire Asset ---")
			assetID := readString(scanner, "Asset ID: ", true)
			reason := readString(scanner, "Retirement Reason: ", true)

			txID, err := retireAsset(assetID, reason)
			if err != nil {
				fmt.Printf("\n[ERROR] Failed to retire asset: %v\n", err)
			} else {
				fmt.Printf("\n[SUCCESS] Asset retired! TxID: %s\n", txID)
			}

		case "13":
			fmt.Println("\n--- Utilization Report ---")
			if err := utilizationReport(); err != nil {
				fmt.Printf("\n[ERROR] Failed to fetch utilization report: %v\n", err)
			}

		case "14":
			fmt.Println("\n--- Compliance Report ---")
			if err := complianceReport(); err != nil {
				fmt.Printf("\n[ERROR] Failed to fetch compliance report: %v\n", err)
			}

		case "15":
			fmt.Println("\nExiting CLI client. Goodbye!")
			return

		default:
			fmt.Println("\n[WARNING] Invalid choice. Please enter a number between 1 and 15.")
		}
	}
}

func readString(scanner *bufio.Scanner, prompt string, required bool) string {
	for {
		fmt.Print(prompt)
		if !scanner.Scan() {
			return ""
		}
		val := strings.TrimSpace(scanner.Text())
		if required && val == "" {
			fmt.Println("Error: input cannot be empty.")
			continue
		}
		return val
	}
}

func readInt(scanner *bufio.Scanner, prompt string, minVal int) int {
	for {
		valStr := readString(scanner, prompt, true)
		val, err := strconv.Atoi(valStr)
		if err != nil {
			fmt.Println("Error: please enter a valid integer.")
			continue
		}
		if val < minVal {
			fmt.Printf("Error: value must be at least %d.\n", minVal)
			continue
		}
		return val
	}
}

func readScore(scanner *bufio.Scanner, prompt string) int {
	for {
		valStr := readString(scanner, prompt, true)
		val, err := strconv.Atoi(valStr)
		if err != nil || val < 1 || val > 5 {
			fmt.Println("Error: please enter an integer between 1 and 5.")
			continue
		}
		return val
	}
}

// REST API structures
type IssueRequest struct {
	DeptID    string `json:"deptID"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Qty       int    `json:"qty"`
	Threshold int    `json:"threshold"`
}

type IssueResponse struct {
	TxID      string    `json:"txID"`
	AssetID   string    `json:"assetID"`
	Timestamp time.Time `json:"timestamp"`
}

type ConsumeRequest struct {
	DeptID  string `json:"deptID"`
	AssetID string `json:"assetID"`
	Qty     int    `json:"qty"`
	Purpose string `json:"purpose"`
}

type ConsumeResponse struct {
	TxID               string `json:"txID"`
	NewQty             int    `json:"newQty"`
	ReplenishTriggered bool   `json:"replenishTriggered"`
}

type TransferRequest struct {
	FromDept string `json:"fromDept"`
	ToDept   string `json:"toDept"`
	AssetID  string `json:"assetID"`
	Qty      int    `json:"qty"`
}

type TransferResponse struct {
	TxID    string `json:"txID"`
	FromQty int    `json:"fromQty"`
	ToQty   int    `json:"toQty"`
}

type ClassifyRequest struct {
	AssetID                string `json:"assetID"`
	BusinessCriticality    int    `json:"businessCriticality"`
	ReplacementCost        int    `json:"replacementCost"`
	ReplacementLeadTime    int    `json:"replacementLeadTime"`
	SafetyComplianceImpact int    `json:"safetyComplianceImpact"`
	RedundancyAvailability int    `json:"redundancyAvailability"`
}

type ClassifyResponse struct {
	TxID             string  `json:"txID"`
	AssetID          string  `json:"assetID"`
	PriorityTier     string  `json:"priorityTier"`
	CriticalityScore float64 `json:"criticalityScore"`
}

type UpdatePriorityRequest struct {
	AssetID      string `json:"assetID"`
	PriorityTier string `json:"priorityTier"`
	Reason       string `json:"reason"`
}

type ScheduleAuditRequest struct {
	AssetID   string `json:"assetID"`
	AuditDate string `json:"auditDate"`
	Scope     string `json:"scope"`
}

type RecordAuditRequest struct {
	AssetID   string `json:"assetID"`
	AuditDate string `json:"auditDate"`
	Result    string `json:"result"`
	Notes     string `json:"notes"`
}

type RetireRequest struct {
	AssetID string `json:"assetID"`
	Reason  string `json:"reason"`
}

type AuditTxResponse struct {
	TxID string `json:"txID"`
}

// Client helper functions calling the REST endpoints
func issueAsset(dept, name, cat string, qty, threshold int) (IssueResponse, error) {
	reqBody, err := json.Marshal(IssueRequest{
		DeptID:    dept,
		Name:      name,
		Category:  cat,
		Qty:       qty,
		Threshold: threshold,
	})
	if err != nil {
		return IssueResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/assets/issue", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return IssueResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return IssueResponse{}, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res IssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return IssueResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return res, nil
}

func consumeStock(dept, assetID string, qty int, purpose string) (ConsumeResponse, error) {
	reqBody, err := json.Marshal(ConsumeRequest{
		DeptID:  dept,
		AssetID: assetID,
		Qty:     qty,
		Purpose: purpose,
	})
	if err != nil {
		return ConsumeResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/assets/consume", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return ConsumeResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ConsumeResponse{}, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res ConsumeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ConsumeResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return res, nil
}

func transferAsset(fromDept, toDept, assetID string, qty int) (TransferResponse, error) {
	reqBody, err := json.Marshal(TransferRequest{
		FromDept: fromDept,
		ToDept:   toDept,
		AssetID:  assetID,
		Qty:      qty,
	})
	if err != nil {
		return TransferResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/assets/transfer", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return TransferResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return TransferResponse{}, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res TransferResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return TransferResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return res, nil
}

func listAssets() error {
	resp, err := http.Get(baseURL + "/api/assets")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var assets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(assets) == 0 {
		fmt.Println("  No assets found.")
		return nil
	}

	for _, a := range assets {
		fmt.Printf("  - AssetID: %s | Dept: %s | Name: %s | Category: %s | Qty: %.0f | Threshold: %.0f\n",
			a["assetId"], a["deptId"], a["name"], a["category"], a["qty"], a["threshold"])
	}
	return nil
}

func getHistory(assetID string) error {
	resp, err := http.Get(baseURL + "/api/assets/" + assetID + "/history")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var txs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(txs) == 0 {
		fmt.Println("  No transaction history found for this asset ID.")
		return nil
	}

	for i, tx := range txs {
		fmt.Printf("  [%d] TxID: %s | Type: %s | Dept: %s | Timestamp: %s\n",
			i, tx["txId"], tx["type"], tx["deptId"], tx["timestamp"])
	}
	return nil
}

func verifyLedger() error {
	resp, err := http.Get(baseURL + "/api/ledger/verify")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("  Ledger Valid:   %v\n", res["valid"])
	if res["brokenAtIndex"] != nil && res["brokenAtIndex"] != -1.0 {
		fmt.Printf("  Broken Index:   %.0f\n", res["brokenAtIndex"])
	}
	return nil
}

func dumpBlocks() error {
	resp, err := http.Get(baseURL + "/api/ledger/blocks")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var blocks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(blocks) == 0 {
		fmt.Println("  No blocks found in the ledger.")
		return nil
	}

	for _, b := range blocks {
		fmt.Printf("  Block #%.0f | Hash: %s | PrevHash: %s | Timestamp: %s\n",
			b["index"], b["hash"], b["prevHash"], b["timestamp"])
		txs, _ := b["transactions"].([]interface{})
		for _, t := range txs {
			tx, _ := t.(map[string]interface{})
			fmt.Printf("    - TxID: %s | Type: %s | Dept: %s\n", tx["txId"], tx["type"], tx["deptId"])
		}
	}
	return nil
}

func postJSONWithRole(url string, body interface{}, role string) (*http.Response, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if role != "" {
		req.Header.Set("X-User-Role", role)
	}
	return http.DefaultClient.Do(req)
}

func getWithRole(url, role string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if role != "" {
		req.Header.Set("X-User-Role", role)
	}
	return http.DefaultClient.Do(req)
}

func classifyAsset(assetID string, bc, rc, lt, sc, ra int) (ClassifyResponse, error) {
	resp, err := postJSONWithRole(baseURL+"/api/assets/classify", ClassifyRequest{
		AssetID:                assetID,
		BusinessCriticality:    bc,
		ReplacementCost:        rc,
		ReplacementLeadTime:    lt,
		SafetyComplianceImpact: sc,
		RedundancyAvailability: ra,
	}, "AI_OPS")
	if err != nil {
		return ClassifyResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ClassifyResponse{}, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res ClassifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ClassifyResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return res, nil
}

func updatePriority(assetID, tier, reason string) (string, error) {
	resp, err := postJSONWithRole(baseURL+"/api/assets/update-priority", UpdatePriorityRequest{
		AssetID:      assetID,
		PriorityTier: tier,
		Reason:       reason,
	}, "ASSET_AUDITOR")
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res AuditTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return res.TxID, nil
}

func scheduleAudit(assetID, auditDate, scope string) (string, error) {
	resp, err := postJSONWithRole(baseURL+"/api/assets/schedule-audit", ScheduleAuditRequest{
		AssetID:   assetID,
		AuditDate: auditDate,
		Scope:     scope,
	}, "ASSET_AUDITOR")
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res AuditTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return res.TxID, nil
}

func recordAudit(assetID, auditDate, result, notes string) (string, error) {
	resp, err := postJSONWithRole(baseURL+"/api/assets/record-audit", RecordAuditRequest{
		AssetID:   assetID,
		AuditDate: auditDate,
		Result:    result,
		Notes:     notes,
	}, "ASSET_AUDITOR")
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res AuditTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return res.TxID, nil
}

func retireAsset(assetID, reason string) (string, error) {
	resp, err := postJSONWithRole(baseURL+"/api/assets/retire", RetireRequest{
		AssetID: assetID,
		Reason:  reason,
	}, "IT_ADMIN")
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var res AuditTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return res.TxID, nil
}

func utilizationReport() error {
	resp, err := getWithRole(baseURL+"/api/reports/utilization", "STORE_MANAGER")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var report map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("  Total Assets: %.0f | Idle Assets: %.0f\n", report["count"], report["idle"])
	assets, _ := report["assets"].([]interface{})
	if len(assets) == 0 {
		fmt.Println("  No assets found.")
		return nil
	}
	for _, a := range assets {
		item, _ := a.(map[string]interface{})
		fmt.Printf("  - %s | %s | Dept: %s | Tier: %s | State: %s | Utilization: %.2f | Idle: %v\n",
			item["assetId"], item["name"], item["deptId"], item["priorityTier"], item["lifecycleState"],
			item["utilizationRate"], item["idle"])
	}
	return nil
}

func complianceReport() error {
	resp, err := getWithRole(baseURL+"/api/reports/compliance", "IT_ADMIN")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var report map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("  Total Assets:   %.0f\n", report["totalAssets"])
	fmt.Printf("  Audit Compliant: %.0f | Overdue: %.0f | Never Audited: %.0f\n",
		report["auditCompliant"], report["auditOverdue"], report["auditNeverRun"])
	return nil
}
