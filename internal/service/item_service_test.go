package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
	"warehouse-management-api/internal/config"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/queue"
	"warehouse-management-api/internal/repository"
	"warehouse-management-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateStock_Logic(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	mockProducer := new(queue.MockEmailProducer)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.AppConfig{EnableLowStockAlert: true, LowStockAlertEmail: "admin@test.com"}

	svc := service.NewItemService(mockRepo, mockProducer, logger, cfg)

	tests := []struct {
		name        string
		oldStock    int
		newStock    int
		expectEmail bool
	}{
		{
			name:        "Stay Above Threshold (10 to 6)",
			oldStock:    10,
			newStock:    6,
			expectEmail: false,
		},
		{
			name:        "Cross Threshold (10 to 3)",
			oldStock:    10,
			newStock:    3,
			expectEmail: true,
		},
		{
			name:        "Already Below Threshold (4 to 2)",
			oldStock:    4,
			newStock:    2,
			expectEmail: false,
		},
	}

	// Gunakan channel untuk sinkronisasi
	done := make(chan bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemID := 1
			req := entity.UpdateStockRequest{ItemID: itemID, Quantity: 1}

			// 1. Mock FindByID (Sebelum Update)
			mockRepo.On("FindByID", itemID).Return(&entity.Item{ID: itemID, Name: "Barang Test", Stock: tt.oldStock}, nil).Once()

			// 2. Mock UpdateStock
			mockRepo.On("UpdateStock", req, 1).Return(nil).Once()

			// 3. Mock FindByID (Sesudah Update)
			mockRepo.On("FindByID", itemID).Return(&entity.Item{ID: itemID, Name: "Barang Test", Stock: tt.newStock}, nil).Once()

			// 4. Mock Producer (Hanya jika expectEmail = true)
			if tt.expectEmail {
				mockProducer.On("PushEmailJob", mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
					// Sinyalkan bahwa producer sudah dipanggil
					done <- true
				})
			}

			// Eksekusi
			err := svc.UpdateStock(context.Background(), req, 1)

			// Assert
			assert.NoError(t, err)

			// Jika kita ekspektasi email, tunggu sampai channel 'done' dikirim
			if tt.expectEmail {
				select {
				case <-done:
					// Berhasil dipanggil
				case <-time.After(1 * time.Second):
					t.Fatal("Timeout: PushEmailJob tidak dipanggil tepat waktu")
				}
			}

			mockRepo.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}
