package service

import (
	"context"
	"log/slog"
	"time"
	pb "warehouse-management-api/internal/pb/audit"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuditClientInterface interface {
	LogActivity(req *pb.AuditRequest)
}

type auditClient struct {
	client    pb.AuditServiceClient
	jwtSecret string
	logger    *slog.Logger
}

func NewAuditClient(conn *grpc.ClientConn, jwtSecret string, logger *slog.Logger) AuditClientInterface {
	return &auditClient{
		client:    pb.NewAuditServiceClient(conn),
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// LogActivity mengirim log ke audit service menggunakan Goroutine (Fire-and-Forget)
func (c *auditClient) LogActivity(req *pb.AuditRequest) {
	// Jalankan di background (Asynchronous)
	go func() {
		// 1. Generate System Token dengan data dari request untuk Auth ke Audit Service
		token, err := c.generateSystemToken(req.Username, req.WarehouseId)
		if err != nil {
			c.logger.Error("Audit Client: Failed to generate token", "error", err)
			return
		}

		// 2. Buat Context dengan Metadata Auth
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		md := metadata.Pairs("authorization", "Bearer "+token)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// 3. Panggil gRPC
		_, err = c.client.LogActivity(ctx, req)
		if err != nil {
			c.logger.Error("Audit Client: Failed to send log", "error", err)
		}
	}()
}

func (c *auditClient) generateSystemToken(username, warehouseID string) (string, error) {
	claims := jwt.MapClaims{
		"username":     username,
		"warehouse_id": warehouseID,
		"iss":          "warehouse-management-api",
		"exp":          time.Now().Add(1 * time.Minute).Unix(), // Token pendek umur
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.jwtSecret))
}
