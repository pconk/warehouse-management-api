package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type LoggerInterceptor struct {
	logger *slog.Logger
}

func NewLoggerInterceptor(logger *slog.Logger) *LoggerInterceptor {
	return &LoggerInterceptor{logger: logger}
}

func (i *LoggerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 1. Cek apakah ada Request ID di metadata (dikirim oleh Gateway/Warehouse)
		md, ok := metadata.FromIncomingContext(ctx)
		var reqID string
		if ok {
			if ids := md.Get("x-request-id"); len(ids) > 0 {
				reqID = ids[0]
			}
		}

		// 2. Jika tidak ada di metadata, baru generate sendiri
		if reqID == "" {
			reqID = generateRequestID()
		}

		// Simpan ID ke context
		ctx = WithRequestID(ctx, reqID)
		start := time.Now()

		// 3. Siapkan wadah untuk log tambahan (username, role, dll)
		collector := &logCollector{fields: make([]slog.Attr, 0)}
		ctx = context.WithValue(ctx, LogFieldsKey, collector)

		// 4. Kirim balik Request ID ke Client lewat Header Response gRPC
		header := metadata.Pairs("x-request-id", reqID)
		grpc.SendHeader(ctx, header)

		// Panggil handler berikutnya (bisa berupa AuthInterceptor atau Handler utama)
		resp, err := handler(ctx, req)

		// Tentukan status, level, dan message
		level := slog.LevelInfo
		statusStr := "OK"
		msg := "gRPC Request Success"

		if err != nil {
			st, _ := status.FromError(err)
			statusStr = st.Code().String()
			level = slog.LevelError
			msg = "gRPC Request Failed"
			// Gunakan helper agar thread-safe
			AddLogFields(ctx, slog.String("error", err.Error()))
		}

		// Log dengan struktur Grouping (trace & grpc)
		i.logger.LogAttrs(ctx, level, msg,
			CreateTraceGroup(reqID, collector),
			slog.Group("grpc",
				slog.String("method", info.FullMethod),
				slog.String("status", statusStr),
				slog.Float64("duration_ms", DurationToMs(time.Since(start))),
			),
		)

		return resp, err
	}
}

func generateRequestID() string {
	b := make([]byte, 8) // 16 karakter hex
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
