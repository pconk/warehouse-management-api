package middleware

import (
	"net/http"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/helper"
)

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value(UserKey).(*entity.User)

			isAllowed := false
			for _, role := range allowedRoles {
				if user.Role == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				helper.SendResponse(w, http.StatusForbidden, "Forbidden", "You don't have permission", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
