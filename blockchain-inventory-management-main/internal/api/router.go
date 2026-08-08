package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRouter registers middleware and routes for Chi HTTP router.
func SetupRouter(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check route
	r.Get("/healthz", h.Healthz)

	// Blockchain Inventory REST Endpoints
	r.Route("/api", func(r chi.Router) {
		r.Route("/assets", func(r chi.Router) {
			r.Get("/", h.ListAssets)
			r.Post("/issue", h.IssueAsset)
			r.Post("/consume", h.ConsumeStock)
			r.Post("/transfer", h.TransferAsset)
			r.Get("/{id}", h.GetAsset)
			r.Get("/{id}/history", h.GetAssetHistory)

			// GenAI-augmented asset prioritization & automation (RBAC protected)
			r.With(RequireRoles(RoleAIOps, RoleAssetAuditor)).Post("/classify", h.ClassifyAsset)
			r.With(RequireRoles(RoleAssetAuditor, RoleITAdmin)).Post("/update-priority", h.UpdatePriority)
			r.With(RequireRoles(RoleAssetAuditor)).Post("/schedule-audit", h.ScheduleAudit)
			r.With(RequireRoles(RoleAssetAuditor)).Post("/record-audit", h.RecordAudit)
			r.With(RequireRoles(RoleITAdmin)).Post("/retire", h.RetireAsset)
		})
		r.Route("/replenish", func(r chi.Router) {
			r.Post("/request", h.RequestReplenishment)
		})
		r.Route("/reports", func(r chi.Router) {
			r.With(RequireRoles(RoleStoreManager, RoleITAdmin)).Get("/utilization", h.UtilizationReport)
			r.With(RequireRoles(RoleITAdmin)).Get("/compliance", h.ComplianceReport)
		})
		r.Route("/assistant", func(r chi.Router) {
			r.With(RequireRoles(RoleDepartmentUser, RoleAssetAuditor, RoleAIOps, RoleITAdmin)).Post("/query", h.AssistantQuery)
		})
		// Admin APIs for runtime control of GenAI agents and models
		r.Route("/admin", func(r chi.Router) {
			r.With(RequireRoles(RoleITAdmin)).Post("/agents/control", h.AgentControl)
		})
		r.Route("/ledger", func(r chi.Router) {
			r.Get("/blocks", h.GetLedgerBlocks)
			r.Get("/verify", h.VerifyLedger)
		})
	})

	return r
}
