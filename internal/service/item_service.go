package service

import (
	"context"
	"fmt"
	"log/slog"
	"warehouse-management-api/internal/config"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/middleware"
	pbAudit "warehouse-management-api/internal/pb/audit"
	"warehouse-management-api/internal/queue"
	"warehouse-management-api/internal/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ItemServiceInterface interface {
	UpdateStock(ctx context.Context, req entity.UpdateStockRequest, user *entity.User) error
	GetAll(ctx context.Context, limit, offset int, filterName string, filterCatID int) ([]entity.Item, int64, error)
	GetAllForExport(ctx context.Context) ([]entity.Item, error)
	GetByID(ctx context.Context, id int) (*entity.Item, error)
	Create(ctx context.Context, item entity.Item) error
	Update(ctx context.Context, id int, item entity.UpdateItemRequest) error
	Delete(ctx context.Context, id int) error
	GetStockLogs(ctx context.Context, limit, offset int, itemID int, logType string) ([]entity.StockLog, int64, error)
}

type itemService struct {
	repo     repository.ItemRepositoryInterface
	producer queue.EmailProducerInterface // Interface untuk push ke Redis
	audit    AuditClientInterface         // Client gRPC Audit
	logger   *slog.Logger
	cfg      *config.AppConfig
}

func NewItemService(r repository.ItemRepositoryInterface, p queue.EmailProducerInterface, a AuditClientInterface, l *slog.Logger, c *config.AppConfig) ItemServiceInterface {
	return &itemService{repo: r, producer: p, audit: a, logger: l, cfg: c}
}

func (s *itemService) UpdateStock(ctx context.Context, req entity.UpdateStockRequest, user *entity.User) error {
	// 1. Ambil data stok SEBELUM update
	// Kita butuh data item (Name, SKU, OldStock) baik untuk Alert maupun Audit
	oldItem, err := s.repo.FindByID(req.ItemID)
	if err != nil || oldItem == nil {
		return fmt.Errorf("item not found")
	}

	// 2. Eksekusi update ke Database
	err = s.repo.UpdateStock(req, user.ID)
	if err != nil {
		return err // Kembalikan error (misal: "insufficient stock")
	}

	// 3. Hitung Stok Akhir (Estimasi) untuk keperluan Log/Alert
	// Idealnya fetch lagi dari DB untuk akurasi, tapi estimasi di sini cukup untuk log
	qtyChange := req.Quantity
	if req.Type == "OUT" {
		qtyChange = -qtyChange
	}
	finalStock := oldItem.Stock + qtyChange

	// 4. Kirim Audit Log (Async via Goroutine di dalam Client Wrapper)
	s.audit.LogActivity(&pbAudit.AuditRequest{
		Username:        user.Username,
		Role:            user.Role,
		WarehouseId:     s.cfg.WarehouseID,
		Action:          fmt.Sprintf("STOCK_%s", req.Type),
		Sku:             oldItem.SKU,
		ProductName:     oldItem.Name,
		QuantityChanged: int32(qtyChange),
		FinalStock:      int32(finalStock),
		Timestamp:       timestamppb.Now(),
		Metadata: map[string]string{
			"source":     "warehouse-api",
			"request_id": middleware.GetRequestID(ctx),
		},
	})

	// 5. Business Logic: Cek stok untuk notifikasi
	if s.cfg.EnableLowStockAlert && oldItem != nil {
		// Cek apakah stok jatuh menembus threshold
		newItem, err := s.repo.FindByID(req.ItemID)
		if err != nil || newItem == nil {
			return nil // Data update masuk tapi gagal ambil info terbaru, abaikan alert
		}

		// Mengambil limit dari config, default ke 5 jika config 0/tidak ada
		limit := s.cfg.LowStockThreshold
		if limit == 0 {
			limit = 5
		}

		if oldItem.Stock >= limit && newItem.Stock < limit {
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

func (s *itemService) GetAll(ctx context.Context, limit, offset int, filterName string, filterCatID int) ([]entity.Item, int64, error) {
	total, err := s.repo.CountAll(filterName, filterCatID)
	if err != nil || total == 0 {
		return nil, 0, err
	}

	items, err := s.repo.FindAll(limit, offset, filterName, filterCatID)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
func (s *itemService) GetAllForExport(ctx context.Context) ([]entity.Item, error) {
	return s.repo.FindAll(10000, 0, "", 0)
}

func (s *itemService) GetByID(ctx context.Context, id int) (*entity.Item, error) {
	return s.repo.FindByID(id)
}

func (s *itemService) Create(ctx context.Context, item entity.Item) error {
	// Bisa tambah validasi bisnis di sini
	return s.repo.Create(item)
}

func (s *itemService) Update(ctx context.Context, id int, item entity.UpdateItemRequest) error {
	return s.repo.Update(id, item)
}

func (s *itemService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(id)
}

func (s *itemService) GetStockLogs(ctx context.Context, limit, offset int, itemID int, logType string) ([]entity.StockLog, int64, error) {
	return s.repo.GetStockLogs(limit, offset, itemID, logType)
}
