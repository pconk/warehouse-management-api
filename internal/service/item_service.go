package service

import (
	"context"
	"fmt"
	"log/slog"
	"warehouse-management-api/internal/config"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/middleware"
	"warehouse-management-api/internal/queue"
	"warehouse-management-api/internal/repository"
)

type ItemServiceInterface interface {
	UpdateStock(ctx context.Context, req entity.UpdateStockRequest, userID int) error
}

type itemService struct {
	repo     repository.ItemRepositoryInterface
	producer queue.EmailProducerInterface // Interface untuk push ke Redis
	logger   *slog.Logger
	cfg      *config.AppConfig
}

func NewItemService(r repository.ItemRepositoryInterface, p queue.EmailProducerInterface, l *slog.Logger, c *config.AppConfig) ItemServiceInterface {
	return &itemService{repo: r, producer: p, logger: l, cfg: c}
}

func (s *itemService) UpdateStock(ctx context.Context, req entity.UpdateStockRequest, userID int) error {
	// 1. Ambil data stok SEBELUM update
	var oldItem *entity.Item
	var err error

	if s.cfg.EnableLowStockAlert {
		oldItem, err = s.repo.FindByID(req.ItemID)
		if err != nil || oldItem == nil {
			return fmt.Errorf("item not found")
		}
	}
	// 2. Eksekusi update ke Database
	err = s.repo.UpdateStock(req, userID)
	if err != nil {
		return err // Kembalikan error (misal: "insufficient stock")
	}

	// 2. Business Logic: Cek stok untuk notifikasi
	if s.cfg.EnableLowStockAlert && oldItem != nil {
		// 3. Ambil data stok SESUDAH update
		newItem, err := s.repo.FindByID(req.ItemID)
		if err != nil || newItem == nil {
			return nil // Data update masuk tapi gagal ambil info terbaru, abaikan alert
		}
		if oldItem.Stock >= 5 && newItem.Stock < 5 {
			s.logger.Info("Stock low, sending email job", "item_id", newItem.ID, "stock", newItem.Stock)

			// Push ke Redis secara async (Goroutine) agar tidak menghambat response API
			go func(itemData entity.Item, reqID string) {
				job := entity.EmailJob{
					ID:      reqID,
					To:      s.cfg.LowStockAlertEmail,
					Subject: fmt.Sprintf("LOW STOCK: %s", newItem.Name),
					Message: fmt.Sprintf("Barang %s tersisa %d unit.", newItem.Name, newItem.Stock),
				}
				// fmt.Println("DEBUG: Memanggil Producer")
				if err := s.producer.PushEmailJob(context.Background(), job); err != nil {
					s.logger.Error("Failed to push email job", "error", err)
				}
			}(*newItem, middleware.GetRequestID(ctx))
		}
	}

	return nil
}
