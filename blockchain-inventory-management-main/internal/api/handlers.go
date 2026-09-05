package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"inventory-chain/internal/genai"
)

type Handlers struct {
	client BlockchainClient
}

func NewHandlers(client BlockchainClient) *Handlers {
	return &Handlers{client: client}
}

// Helper to write JSON error messages
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// Helper to write JSON success messages
func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// 1. POST /api/assets/issue
type IssueRequest struct {
	DeptID    string `json:"deptID,omitempty"`
	DeptId    string `json:"deptId,omitempty"`
	Name      string `json:"name,omitempty"`
	Category  string `json:"category,omitempty"`
	Qty       int    `json:"qty,omitempty"`
	Quantity  int    `json:"quantity,omitempty"`
	Threshold int    `json:"threshold,omitempty"`
}

func (r *IssueRequest) normalize() {
	if r.DeptID == "" {
		r.DeptID = r.DeptId
	}
	if r.Qty == 0 && r.Quantity != 0 {
		r.Qty = r.Quantity
	}
}

type IssueResponse struct {
	TxID      string    `json:"txID"`
	AssetID   string    `json:"assetID"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Handlers) IssueAsset(w http.ResponseWriter, r *http.Request) {
	var req IssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, assetID, timestamp, err := h.client.IssueAsset(req.DeptID, req.Name, req.Category, req.Qty, req.Threshold)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusCreated, IssueResponse{
		TxID:      txID,
		AssetID:   assetID,
		Timestamp: timestamp,
	})
}

// 2. POST /api/assets/consume
type ConsumeRequest struct {
	DeptID   string `json:"deptID,omitempty"`
	DeptId   string `json:"deptId,omitempty"`
	AssetID  string `json:"assetID,omitempty"`
	AssetId  string `json:"assetId,omitempty"`
	Qty      int    `json:"qty,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	Purpose  string `json:"purpose,omitempty"`
}

func (r *ConsumeRequest) normalize() {
	if r.DeptID == "" {
		r.DeptID = r.DeptId
	}
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
	if r.Qty == 0 && r.Quantity != 0 {
		r.Qty = r.Quantity
	}
}

type ConsumeResponse struct {
	TxID               string `json:"txID"`
	NewQty             int    `json:"newQty"`
	ReplenishTriggered bool   `json:"replenishTriggered"`
}

func (h *Handlers) ConsumeStock(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, newQty, replenishTriggered, err := h.client.ConsumeStock(req.DeptID, req.AssetID, req.Qty, req.Purpose)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, ConsumeResponse{
		TxID:               txID,
		NewQty:             newQty,
		ReplenishTriggered: replenishTriggered,
	})
}

// 3. POST /api/assets/transfer
type TransferRequest struct {
	FromDept   string `json:"fromDept,omitempty"`
	FromDeptId string `json:"fromDeptId,omitempty"`
	ToDept     string `json:"toDept,omitempty"`
	ToDeptId   string `json:"toDeptId,omitempty"`
	AssetID    string `json:"assetID,omitempty"`
	AssetId    string `json:"assetId,omitempty"`
	Qty        int    `json:"qty,omitempty"`
	Quantity   int    `json:"quantity,omitempty"`
}

func (r *TransferRequest) normalize() {
	if r.FromDept == "" {
		r.FromDept = r.FromDeptId
	}
	if r.ToDept == "" {
		r.ToDept = r.ToDeptId
	}
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
	if r.Qty == 0 && r.Quantity != 0 {
		r.Qty = r.Quantity
	}
}

type TransferResponse struct {
	TxID    string `json:"txID"`
	FromQty int    `json:"fromQty"`
	ToQty   int    `json:"toQty"`
}

func (h *Handlers) TransferAsset(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	if req.FromDept == "" && req.AssetID != "" {
		if asset, err := h.client.ReadAsset(req.AssetID); err == nil && asset != nil {
			req.FromDept = asset.DeptID
		}
	}

	txID, fromQty, toQty, err := h.client.TransferAsset(req.FromDept, req.ToDept, req.AssetID, req.Qty)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, TransferResponse{
		TxID:    txID,
		FromQty: fromQty,
		ToQty:   toQty,
	})
}

// 4. GET /api/assets/{id}/history
type HistoryTxResponse struct {
	TxID      string          `json:"txId"`
	Type      string          `json:"type"`
	DeptID    string          `json:"deptId"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

func (h *Handlers) GetAssetHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	history, err := h.client.GetAssetHistory(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var txs []HistoryTxResponse
	for i, entry := range history {
		// Fabric's GetHistoryForKey returns newest-first and does not expose which
		// chaincode function was invoked, so the type is inferred from what actually
		// changed between this entry and the next-older one.
		var prev *ClientAsset
		if i+1 < len(history) {
			prev = history[i+1].Value
		}
		isOldest := i == len(history)-1
		txType := inferHistoryType(entry, prev, isOldest)

		deptID := ""
		var payload []byte
		if entry.Value != nil {
			deptID = entry.Value.DeptID
			payload, _ = json.Marshal(entry.Value)
		}

		txs = append(txs, HistoryTxResponse{
			TxID:      entry.TxID,
			Type:      txType,
			DeptID:    deptID,
			Timestamp: entry.Timestamp,
			Payload:   payload,
		})
	}

	if txs == nil {
		txs = []HistoryTxResponse{}
	}
	writeJSONResponse(w, http.StatusOK, txs)
}

// inferHistoryType derives a human-readable transaction type by diffing an
// asset snapshot against the state it replaced, since Fabric's key history
// only exposes state values, not the invoked function name.
func inferHistoryType(entry ClientHistoryEntry, prev *ClientAsset, isOldest bool) string {
	if entry.IsDelete {
		return "DELETE"
	}
	if isOldest || prev == nil || entry.Value == nil {
		return "ISSUE"
	}
	v := entry.Value
	switch {
	case v.LifecycleState == genai.LifecycleRetired && prev.LifecycleState != genai.LifecycleRetired:
		return "RETIRE"
	case v.Qty < prev.Qty:
		// Both ConsumeStock and the source side of TransferAsset decrement Qty;
		// Fabric's history alone can't tell them apart.
		return "STOCK_DECREASE"
	case v.LastAuditDate != prev.LastAuditDate:
		return "AUDIT"
	case v.PriorityTier != prev.PriorityTier || v.CriticalityScore != prev.CriticalityScore:
		return "CLASSIFY"
	case v.WarrantyExpiry != prev.WarrantyExpiry || v.AMCExpiry != prev.AMCExpiry:
		return "DOCUMENT_UPDATE"
	default:
		return "UPDATE"
	}
}

// 5. GET /api/assets/{id}
func (h *Handlers) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, err := h.client.ReadAsset(id)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("asset not found: %s", id), http.StatusNotFound)
		return
	}
	writeJSONResponse(w, http.StatusOK, asset)
}

// 6. GET /api/assets
func (h *Handlers) ListAssets(w http.ResponseWriter, r *http.Request) {
	deptFilter := r.URL.Query().Get("dept")
	assets, err := h.client.QueryAssetsByDept(deptFilter)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if assets == nil {
		assets = []*ClientAsset{}
	}
	writeJSONResponse(w, http.StatusOK, assets)
}

// 7. POST /api/replenish/request
type ReplenishRequest struct {
	AssetID  string `json:"assetID,omitempty"`
	AssetId  string `json:"assetId,omitempty"`
	Qty      int    `json:"qty,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	Urgency  string `json:"urgency,omitempty"`
}

func (r *ReplenishRequest) normalize() {
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
	if r.Qty == 0 && r.Quantity != 0 {
		r.Qty = r.Quantity
	}
}

type ReplenishResponse struct {
	TxID string `json:"txID"`
}

func (h *Handlers) RequestReplenishment(w http.ResponseWriter, r *http.Request) {
	var req ReplenishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, err := h.client.RequestReplenishment(req.AssetID, req.Qty, req.Urgency)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, ReplenishResponse{TxID: txID})
}

// 8. GET /api/ledger/blocks (Returns static list in production Fabric context)
func (h *Handlers) GetLedgerBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := h.client.GetLedgerBlocks()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, blocks)
}

// 9. GET /api/ledger/verify
type VerifyResponse struct {
	Valid         bool `json:"valid"`
	BrokenAtIndex *int `json:"brokenAtIndex"`
}

func (h *Handlers) VerifyLedger(w http.ResponseWriter, r *http.Request) {
	valid, brokenIndex, err := h.client.VerifyLedger()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, VerifyResponse{
		Valid:         valid,
		BrokenAtIndex: brokenIndex,
	})
}

// 10. GET /healthz
func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// 11. POST /api/assets/classify
type ClassifyRequest struct {
	AssetID                string `json:"assetID,omitempty"`
	AssetId                string `json:"assetId,omitempty"`
	BusinessCriticality    int    `json:"businessCriticality,omitempty"`
	ReplacementCost        int    `json:"replacementCost,omitempty"`
	ReplacementLeadTime    int    `json:"replacementLeadTime,omitempty"`
	SafetyComplianceImpact int    `json:"safetyComplianceImpact,omitempty"`
	RedundancyAvailability int    `json:"redundancyAvailability,omitempty"`
}

func (r *ClassifyRequest) normalize() {
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
}

type ClassifyResponse struct {
	TxID             string  `json:"txID"`
	AssetID          string  `json:"assetID"`
	PriorityTier     string  `json:"priorityTier"`
	CriticalityScore float64 `json:"criticalityScore"`
}

func (h *Handlers) ClassifyAsset(w http.ResponseWriter, r *http.Request) {
	var req ClassifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	scores := genai.PriorityScores{
		BusinessCriticality:    req.BusinessCriticality,
		ReplacementCost:        req.ReplacementCost,
		ReplacementLeadTime:    req.ReplacementLeadTime,
		SafetyComplianceImpact: req.SafetyComplianceImpact,
		RedundancyAvailability: req.RedundancyAvailability,
	}

	if !scores.Valid() {
		asset, err := h.client.ReadAsset(req.AssetID)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("asset %s not found: %v", req.AssetID, err), http.StatusNotFound)
			return
		}
		scores = genai.ScorePriority(asset.Name, asset.Category)
	}

	txID, tier, score, err := h.client.ClassifyPriority(req.AssetID, scores)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, ClassifyResponse{
		TxID:             txID,
		AssetID:          req.AssetID,
		PriorityTier:     tier,
		CriticalityScore: score,
	})
}

// 12. POST /api/assets/update-priority
type UpdatePriorityRequest struct {
	AssetID      string `json:"assetID,omitempty"`
	AssetId      string `json:"assetId,omitempty"`
	PriorityTier string `json:"priorityTier,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

func (r *UpdatePriorityRequest) normalize() {
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
}

type UpdatePriorityResponse struct {
	TxID string `json:"txID"`
}

func (h *Handlers) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	var req UpdatePriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, err := h.client.UpdatePriorityTier(req.AssetID, req.PriorityTier, req.Reason)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, UpdatePriorityResponse{TxID: txID})
}

// 13. POST /api/assets/schedule-audit
type ScheduleAuditRequest struct {
	AssetID       string `json:"assetID,omitempty"`
	AssetId       string `json:"assetId,omitempty"`
	AuditDate     string `json:"auditDate,omitempty"`
	ScheduledDate string `json:"scheduledDate,omitempty"`
	Scope         string `json:"scope,omitempty"`
}

func (r *ScheduleAuditRequest) normalize() {
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
	if r.AuditDate == "" {
		r.AuditDate = r.ScheduledDate
	}
}

type AuditTxResponse struct {
	TxID string `json:"txID"`
}

func (h *Handlers) ScheduleAudit(w http.ResponseWriter, r *http.Request) {
	var req ScheduleAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, err := h.client.ScheduleAudit(req.AssetID, req.AuditDate, req.Scope)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, AuditTxResponse{TxID: txID})
}

// 14. POST /api/assets/record-audit
type RecordAuditRequest struct {
	AssetID       string `json:"assetID,omitempty"`
	AssetId       string `json:"assetId,omitempty"`
	AuditDate     string `json:"auditDate,omitempty"`
	ScheduledDate string `json:"scheduledDate,omitempty"`
	Result        string `json:"result,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

func (r *RecordAuditRequest) normalize() {
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
	if r.AuditDate == "" {
		r.AuditDate = r.ScheduledDate
	}
}

func (h *Handlers) RecordAudit(w http.ResponseWriter, r *http.Request) {
	var req RecordAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, err := h.client.RecordAuditResult(req.AssetID, req.AuditDate, req.Result, req.Notes)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, AuditTxResponse{TxID: txID})
}

// 15. POST /api/assets/retire
type RetireRequest struct {
	AssetID string `json:"assetID,omitempty"`
	AssetId string `json:"assetId,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

func (r *RetireRequest) normalize() {
	if r.AssetID == "" {
		r.AssetID = r.AssetId
	}
}

func (h *Handlers) RetireAsset(w http.ResponseWriter, r *http.Request) {
	var req RetireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.normalize()

	txID, err := h.client.RetireAsset(req.AssetID, req.Reason)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, http.StatusOK, AuditTxResponse{TxID: txID})
}

// 16. POST /api/assistant/query
type AssistantQueryRequest struct {
	Query string `json:"query"`
}

type AssistantQueryResponse struct {
	Answer string `json:"answer"`
}

func (h *Handlers) AssistantQuery(w http.ResponseWriter, r *http.Request) {
	var req AssistantQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	answer, err := h.client.AssistantQuery(req.Query)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, AssistantQueryResponse{Answer: answer})
}

// 17. GET /api/reports/utilization
type UtilizationItem struct {
	AssetID         string  `json:"assetId"`
	Name            string  `json:"name"`
	DeptID          string  `json:"deptId"`
	PriorityTier    string  `json:"priorityTier"`
	LifecycleState  string  `json:"lifecycleState"`
	UtilizationRate float64 `json:"utilizationRate"`
	Idle            bool    `json:"idle"`
}

func (h *Handlers) UtilizationReport(w http.ResponseWriter, r *http.Request) {
	assets, err := h.client.QueryAssetsByDept("")
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var items []UtilizationItem
	idleThreshold := 0.10
	for _, a := range assets {
		u, err := h.client.ComputeUtilization(a.AssetID)
		if err != nil {
			u = 0
		}
		idle := u < idleThreshold && a.LifecycleState != genai.LifecycleRetired
		items = append(items, UtilizationItem{
			AssetID:         a.AssetID,
			Name:            a.Name,
			DeptID:          a.DeptID,
			PriorityTier:    a.PriorityTier,
			LifecycleState:  a.LifecycleState,
			UtilizationRate: u,
			Idle:            idle,
		})
	}

	if items == nil {
		items = []UtilizationItem{}
	}
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"count":  len(items),
		"idle":   countIdle(items),
		"assets": items,
	})
}

func countIdle(items []UtilizationItem) int {
	n := 0
	for _, it := range items {
		if it.Idle {
			n++
		}
	}
	return n
}

// 17. GET /api/reports/compliance
type ComplianceSummary struct {
	TotalAssets      int                 `json:"totalAssets"`
	ByTier           map[string]int      `json:"byTier"`
	ByLifecycleState map[string]int      `json:"byLifecycleState"`
	AuditCompliant   int                 `json:"auditCompliant"`
	AuditOverdue     int                 `json:"auditOverdue"`
	AuditNeverRun    int                 `json:"auditNeverRun"`
	OverdueByTier    map[string][]string `json:"overdueByTier"`
}

func auditCadenceDays(tier string) int {
	switch tier {
	case "P1":
		return 30
	case "P2":
		return 180
	default:
		return 365
	}
}

func (h *Handlers) ComplianceReport(w http.ResponseWriter, r *http.Request) {
	assets, err := h.client.QueryAssetsByDept("")
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := ComplianceSummary{
		ByTier:           make(map[string]int),
		ByLifecycleState: make(map[string]int),
		OverdueByTier:    make(map[string][]string),
	}

	today := time.Now()
	for _, a := range assets {
		summary.TotalAssets++
		if a.PriorityTier != "" {
			summary.ByTier[a.PriorityTier]++
		}
		state := a.LifecycleState
		if state == "" {
			state = genai.LifecycleActive
		}
		summary.ByLifecycleState[state]++

		if state == genai.LifecycleRetired {
			continue
		}

		if a.LastAuditDate == "" {
			summary.AuditNeverRun++
			summary.OverdueByTier[a.PriorityTier] = append(summary.OverdueByTier[a.PriorityTier], a.AssetID)
			continue
		}
		auditDate, err := time.Parse("2006-01-02", a.LastAuditDate)
		if err != nil {
			summary.AuditNeverRun++
			continue
		}
		if today.Sub(auditDate) > time.Duration(auditCadenceDays(a.PriorityTier))*24*time.Hour {
			summary.AuditOverdue++
			summary.OverdueByTier[a.PriorityTier] = append(summary.OverdueByTier[a.PriorityTier], a.AssetID)
		} else {
			summary.AuditCompliant++
		}
	}

	writeJSONResponse(w, http.StatusOK, summary)
}
