package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRouter registers middleware and routes for Chi HTTP router.
func SetupRouter(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable CORS for frontend clients
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Role")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

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

	// Serve frontend static files if built
	distDir := "./frontend/dist"
	if _, err := os.Stat(distDir); err == nil {
		fs := http.FileServer(http.Dir(distDir))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/healthz") {
				http.NotFound(w, r)
				return
			}
			filePath := filepath.Join(distDir, filepath.Clean(r.URL.Path))
			if info, err := os.Stat(filePath); err != nil || info.IsDir() {
				http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		})
	}

	return r
}
