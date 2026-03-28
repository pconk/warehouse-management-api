package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	LogFieldsKey contextKey     = "log_fields"
	RequestIDKey authContextKey = "requestID"
)

// AddLogFields memungkinkan middleware lain menambahkan field ke log utama
func AddLogFields(ctx context.Context, fields ...slog.Attr) {
	if v, ok := ctx.Value(LogFieldsKey).(*[]slog.Attr); ok {
		*v = append(*v, fields...)
	}
}

func AddUserToLog(ctx context.Context, userID int64, username, role string) {
	AddLogFields(ctx,
		slog.String("user_id", fmt.Sprintf("%d", userID)),
		slog.String("username", username),
		slog.String("role", role))
}

// CreateTraceGroup menggabungkan request_id dan field tambahan ke dalam satu grup log trace
func CreateTraceGroup(requestID string, extraFields []slog.Attr) slog.Attr {
	traceAttrs := make([]any, 0, len(extraFields)+1)
	traceAttrs = append(traceAttrs, slog.String("request_id", requestID))
	for _, attr := range extraFields {
		traceAttrs = append(traceAttrs, attr)
	}
	return slog.Group("trace", traceAttrs...)
}

// GetRequestID adalah helper untuk mengambil ID di dalam Handler
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// DurationToMs mengonversi time.Duration menjadi float64 milidetik
func DurationToMs(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
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

			var extraFields []slog.Attr
			ctx = context.WithValue(ctx, LogFieldsKey, &extraFields)

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
			// userID := wrapped.Header().Get("X-Internal-User-ID")
			// username := wrapped.Header().Get("X-Internal-Username")
			// role := wrapped.Header().Get("X-Internal-Role")

			// wrapped.Header().Del("X-Internal-User-ID")
			// wrapped.Header().Del("X-Internal-Username")
			// wrapped.Header().Del("X-Internal-Role")

			// 6. Log Terstruktur dengan slog + Request ID (Menggunakan AddLogFields dari middleware Auth)
			logger.LogAttrs(r.Context(), level, "HTTP Request",
				CreateTraceGroup(requestID, extraFields),
				slog.Group("http",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", wrapped.statusCode),
					slog.String("ip", r.RemoteAddr),
					slog.Float64("duration_ms", DurationToMs(time.Since(start))),
				),
			)

		})
	}
}
