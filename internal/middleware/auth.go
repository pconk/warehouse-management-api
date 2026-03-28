package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/helper"

	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type authContextKey string

const UserKey authContextKey = "user_context_key"
const TokenKey authContextKey = "token_context_key"

func GetUser(ctx context.Context) *entity.User {
	// Pastikan casting ke *entity.User (pakai bintang)
	user, ok := ctx.Value(UserKey).(*entity.User)
	if !ok {
		return nil
	}
	return user
}

func GetToken(ctx context.Context) string {
	token, ok := ctx.Value(TokenKey).(string)
	if !ok {
		return ""
	}
	return token
}

func AuthMiddleware(logger *slog.Logger, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Ambil token dari header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				helper.SendResponse(w, http.StatusUnauthorized, "Unauthorized", "Authorization header is required", nil)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// 2. Parse dan Validasi Token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				helper.SendResponse(w, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token", nil)
				return
			}

			// 3. Mapping Claims ke entity.User
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				helper.SendResponse(w, http.StatusUnauthorized, "Unauthorized", "Invalid token claims", nil)
				return
			}

			var userID int64
			if id, ok := claims["user_id"]; ok {
				switch v := id.(type) {
				case float64:
					userID = int64(v)
				case string:
					fmt.Sscanf(v, "%d", &userID)
				}
			}

			usernameClaim, _ := claims["username"].(string)
			roleClaim, _ := claims["role"].(string)

			if userID == 0 || usernameClaim == "" || roleClaim == "" {
				helper.SendResponse(w, http.StatusUnauthorized, "Unauthorized", "Token missing or invalid user fields", nil)
				return
			}

			user := &entity.User{
				ID:       userID,
				Username: usernameClaim,
				Role:     roleClaim,
			}
			w.Header().Set("X-Internal-User-ID", fmt.Sprintf("%d", userID))
			w.Header().Set("X-Internal-Username", usernameClaim)
			w.Header().Set("X-Internal-Role", roleClaim)

			// 4. Simpan User dan Token mentah ke Context
			ctx := context.WithValue(r.Context(), UserKey, user)
			ctx = context.WithValue(ctx, TokenKey, tokenString)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
