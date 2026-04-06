package handler

import (
	"context"
	"log/slog"
	"strings"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/middleware"

	pb_warehouse "warehouse-management-api/internal/pb/warehouse" // Import dengan alias

	"warehouse-management-api/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WarehouseGRPCHandler struct {
	pb_warehouse.UnimplementedWarehouseServiceServer

	itemService service.ItemServiceInterface
	logger      *slog.Logger
}

func NewWarehouseGRPCHandler(s service.ItemServiceInterface, l *slog.Logger) *WarehouseGRPCHandler {
	return &WarehouseGRPCHandler{
		itemService: s,
		logger:      l,
	}
}

func (h *WarehouseGRPCHandler) IncreaseStock(ctx context.Context, req *pb_warehouse.StockUpdateRequest) (*pb_warehouse.StockUpdateResponse, error) {
	// 1. Validasi input dari gRPC request menggunakan generated Validate() method
	if err := req.Validate(); err != nil {
		h.logger.Warn("gRPC: IncreaseStock - Validation failed", "error", err.Error(), "request_id", middleware.GetRequestID(ctx))
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Log request setelah validasi dasar

	h.logger.Info("gRPC: IncreaseStock request received", "total_items", len(req.Items))

	user := middleware.GetUser(ctx)
	if user == nil {
		h.logger.Error("gRPC: IncreaseStock - User not found in context", "request_id", middleware.GetRequestID(ctx))
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	for _, item := range req.Items {
		h.logger.Debug("Updating stock for item", "item_id", item.ItemId, "quantity", item.Quantity)
		updateReq := entity.UpdateStockRequest{
			ItemID:   int(item.ItemId),
			Type:     "IN",
			Quantity: int(item.Quantity),
			Reason:   "PO Received",
		}

		if err := h.itemService.UpdateStock(ctx, updateReq, user); err != nil {
			h.logger.Error("Failed to update stock via gRPC", "item_id", item.ItemId, "error", err)
			// Terjemahkan error dari service ke gRPC status codes yang sesuai.
			if strings.Contains(err.Error(), "item not found") {
				return nil, status.Errorf(codes.NotFound, "item with ID %d not found", item.ItemId)
			}
			return nil, status.Errorf(codes.Internal, "failed to update stock for item %d: %v", item.ItemId, err)
		}
	}

	return &pb_warehouse.StockUpdateResponse{Success: true}, nil
}
