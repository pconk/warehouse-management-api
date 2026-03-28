package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey authContextKey = "requestID"

// GetRequestID adalah helper untuk mengambil ID di dalam Handler
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// responseWriter wrapper untuk menangkap status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 1. Generate Request ID (Correlation ID)
			requestID := uuid.New().String()

			// 2. Masukkan Request ID ke Header Response (opsional, untuk debug client)
			w.Header().Set("X-Request-ID", requestID)

			// 3. Masukkan Request ID ke Context agar bisa dipakai di Handler/DB
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// 4. Wrap ResponseWriter
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// 5. Jalankan Handler berikutnya
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Tentukan level log berdasarkan status code
			level := slog.LevelInfo
			if wrapped.statusCode >= 400 && wrapped.statusCode < 500 {
				level = slog.LevelWarn
			} else if wrapped.statusCode >= 500 {
				level = slog.LevelError
			}

			// Gunakan variabel default
			userID := wrapped.Header().Get("X-Internal-User-ID")
			username := wrapped.Header().Get("X-Internal-Username")
			role := wrapped.Header().Get("X-Internal-Role")

			wrapped.Header().Del("X-Internal-User-ID")
			wrapped.Header().Del("X-Internal-Username")
			wrapped.Header().Del("X-Internal-Role")

			// Log dengan struktur yang rapi (Nesting)
			logger.LogAttrs(r.Context(), level, "HTTP Request",
				slog.Group("trace",
					slog.String("request_id", requestID),
					slog.String("user_id", userID),
					slog.String("username", username),
					slog.String("role", role),
				),
				slog.Group("http",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", wrapped.statusCode),
					slog.String("ip", r.RemoteAddr),
					slog.Duration("duration", time.Since(start)),
				),
			)

		})
	}
}
