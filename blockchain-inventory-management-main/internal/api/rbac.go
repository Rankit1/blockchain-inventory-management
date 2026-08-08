package api

import (
	"net/http"
	"strings"
)

const (
	RoleAssetAuditor   = "ASSET_AUDITOR"
	RoleAIOps          = "AI_OPS"
	RoleStoreManager   = "STORE_MANAGER"
	RoleDepartmentUser = "DEPARTMENT_USER"
	RoleITAdmin        = "IT_ADMIN"
	RoleSystemAdmin    = "SYSTEM_ADMIN"
)

// RequireRoles returns middleware that enforces role-based access control.
// The caller's role is read from the X-User-Role header (the application gateway
// should populate this from the validated JWT role claim). SYSTEM_ADMIN is always allowed.
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles)+1)
	for _, r := range roles {
		allowed[r] = true
	}
	allowed[RoleSystemAdmin] = true

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := strings.TrimSpace(r.Header.Get("X-User-Role"))
			if !allowed[role] {
				writeJSONError(w, "forbidden: insufficient role", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
