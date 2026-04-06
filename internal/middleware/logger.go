package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const (
	LogFieldsKey contextKey = "log_fields"
	RequestIDKey contextKey = "request_id"
)

// logCollector membungkus slice attr agar thread-safe
type logCollector struct {
	mu     sync.RWMutex
	fields []slog.Attr
}

// AddLogFields memungkinkan middleware lain menambahkan field ke log utama
func AddLogFields(ctx context.Context, fields ...slog.Attr) {
	if v, ok := ctx.Value(LogFieldsKey).(*logCollector); ok {
		v.mu.Lock()
		defer v.mu.Unlock()
		v.fields = append(v.fields, fields...)
	}
}

func AddUserToLog(ctx context.Context, userID int64, username, role string) {
	AddLogFields(ctx,
		slog.String("user_id", fmt.Sprintf("%d", userID)),
		slog.String("username", username),
		slog.String("role", role))
}

// CreateTraceGroup menggabungkan request_id dan field tambahan ke dalam satu grup log trace
func CreateTraceGroup(requestID string, collector *logCollector) slog.Attr {
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	traceAttrs := make([]any, 0, len(collector.fields)+1)
	traceAttrs = append(traceAttrs, slog.String("request_id", requestID))
	for _, attr := range collector.fields {
		traceAttrs = append(traceAttrs, attr)
	}
	return slog.Group("trace", traceAttrs...)
}

// WithRequestID menyuntikkan request ID ke dalam context secara manual
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID adalah helper untuk mengambil ID di dalam Handler
func GetRequestID(ctx context.Context) string {
	// Cek apakah ada di key internal kita (untuk context asinkron)
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	// Fallback ke chi middleware (untuk context request HTTP)
	return chimiddleware.GetReqID(ctx)
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

			// 1. Ambil Request ID yang sudah digenerate oleh chi.middleware.RequestID
			requestID := GetRequestID(r.Context())

			collector := &logCollector{fields: make([]slog.Attr, 0)}
			ctx := context.WithValue(r.Context(), LogFieldsKey, collector)

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

			// 6. Log Terstruktur dengan slog + Request ID (Menggunakan AddLogFields dari middleware Auth)
			logger.LogAttrs(r.Context(), level, "HTTP Request",
				CreateTraceGroup(requestID, collector),
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
