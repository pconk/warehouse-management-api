package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"warehouse-management-api/internal/config"
	"warehouse-management-api/internal/entity"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtSecret string
}

func NewAuthInterceptor(cfg *config.AppConfig) *AuthInterceptor {
	return &AuthInterceptor{
		jwtSecret: cfg.JWTSecret,
	}
}

// Unary mengembalikan interceptor untuk validasi JWT pada setiap request
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 1. Ambil Metadata (Header)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		// 2. Ambil Header Authorization
		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		// Format: "Bearer <token>"
		tokenString := strings.Replace(values[0], "Bearer ", "", 1)

		// 3. Parse & Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(i.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// 4. Extract Data dari Claims (Username & WarehouseID)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Ambil UserID dengan aman (bisa string atau float64 dari JWT)
			var userID int64
			if id, ok := claims["user_id"]; ok {
				switch v := id.(type) {
				case float64:
					userID = int64(v)
				case string:
					parsedID, err := strconv.ParseInt(v, 10, 64)
					if err == nil {
						userID = parsedID
					}
				}
			}

			username, _ := claims["username"].(string)
			role, _ := claims["role"].(string)

			// Buat objek User
			user := &entity.User{
				ID:       userID,
				Username: username,
				Role:     role,
			}

			// Tambahkan info user ke Log LoggerInterceptor
			AddUserToLog(ctx, user.ID, user.Username, user.Role)
			// Simpan user dan token ke context agar bisa diakses oleh handler
			ctx = context.WithValue(ctx, UserKey, user)
			ctx = WithToken(ctx, tokenString)
		}

		// Jika valid, lanjutkan ke handler utama
		return handler(ctx, req)
	}
}
