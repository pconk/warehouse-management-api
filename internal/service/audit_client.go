package service

import (
	"context"
	"log/slog"
	"time"
	"warehouse-management-api/internal/middleware"
	pb "warehouse-management-api/internal/pb/audit"

	"google.golang.org/grpc"
)

type AuditClientInterface interface {
	LogActivity(ctx context.Context, req *pb.AuditRequest)
}

type auditClient struct {
	client pb.AuditServiceClient
	logger *slog.Logger
}

func NewAuditClient(conn *grpc.ClientConn, logger *slog.Logger) AuditClientInterface {
	return &auditClient{
		client: pb.NewAuditServiceClient(conn),
		logger: logger,
	}
}

// LogActivity mengirim log ke audit service menggunakan Goroutine (Fire-and-Forget)
func (c *auditClient) LogActivity(ctx context.Context, req *pb.AuditRequest) {
	// Ambil requestID dari context asli sebelum goroutine dimulai
	reqID := middleware.GetRequestID(ctx)
	token := middleware.GetToken(ctx)

	// Jalankan di background (Asynchronous)
	go func() {
		// Gunakan context.Background() agar tidak ikut mati saat request utama selesai
		// Suntikkan kembali reqID dan token agar interceptor gRPC bisa membacanya
		asyncCtx := middleware.WithRequestID(context.Background(), reqID)
		asyncCtx = middleware.WithToken(asyncCtx, token)

		asyncCtx, cancel := context.WithTimeout(asyncCtx, 5*time.Second)
		defer cancel()

		// 3. Panggil gRPC
		_, err := c.client.LogActivity(asyncCtx, req)
		if err != nil {
			c.logger.Error("Audit Client: Failed to send log", "error", err)
		}
	}()
}
