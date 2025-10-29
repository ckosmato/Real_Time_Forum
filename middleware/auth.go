package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"real-time-forum/services"
	"real-time-forum/utils"
)

type AuthMiddleware struct{
	UserService services.UserService
}

func NewAuthMiddleware(us services.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		UserService: us,
	}
}

func (m *AuthMiddleware) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")

		if sessionID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}
		// Get user from session
		user, err := m.UserService.GetUserBySessionID(r.Context(), sessionID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), utils.ContextUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
	

