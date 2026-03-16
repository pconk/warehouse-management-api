package middleware

import (
	"context"
	"log/slog"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/helper"
	"warehouse-management-api/internal/repository"

	"net/http"
)

const UserKey contextKey = "user_context_key"

func GetUser(ctx context.Context) *entity.User {
	// Pastikan casting ke *entity.User (pakai bintang)
	user, ok := ctx.Value(UserKey).(*entity.User)
	if !ok {
		return nil
	}
	return user
}

func AuthMiddleware(repo *repository.UserRepository, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Ambil API Key dari Header
			apiKey := r.Header.Get("X-API-KEY")

			if apiKey == "" {
				helper.SendResponse(w, http.StatusUnauthorized, "Unauthorized", "API Key is missing", nil)
				return
			}

			// Query ke DB untuk cek user & role
			user, err := repo.FindByApiKey(apiKey)

			logger.Info("AuthMiddleware", "user", user)
			if err != nil {
				logger.Warn("Unauthorized access attempt", "api_key", apiKey, "error", err)
				helper.SendResponse(w, http.StatusUnauthorized, "Unauthorized", "Invalid API Key", nil)
				return
			}

			// Simpan data user ke context supaya bisa dipake di handler (misal buat log siapa yg update)
			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
