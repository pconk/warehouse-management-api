package service

import (
	"context"
	"log/slog"
	"time"
	pb "warehouse-management-api/internal/pb/audit"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuditClientInterface interface {
	LogActivity(req *pb.AuditRequest, token string)
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
func (c *auditClient) LogActivity(req *pb.AuditRequest, token string) {
	// Jalankan di background (Asynchronous)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		md := metadata.Pairs("authorization", "Bearer "+token)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// 3. Panggil gRPC
		_, err := c.client.LogActivity(ctx, req)
		if err != nil {
			c.logger.Error("Audit Client: Failed to send log", "error", err)
		}
	}()
}
