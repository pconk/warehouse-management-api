package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"
	"warehouse-management-api/internal/config"
	"warehouse-management-api/internal/entity"
	audit "warehouse-management-api/internal/pb/audit"
	"warehouse-management-api/internal/queue"
	"warehouse-management-api/internal/repository"
	"warehouse-management-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditClient
type MockAuditClient struct {
	mock.Mock
}

func (m *MockAuditClient) LogActivity(req *audit.AuditRequest) {
	m.Called(req)
}

// --- SETUP HELPER ---
// Helper agar tidak mengulang kode inisialisasi di setiap test function
func setupItemService(t *testing.T) (service.ItemServiceInterface, *repository.MockItemRepo, *queue.MockEmailProducer, *MockAuditClient) {
	mockRepo := new(repository.MockItemRepo)
	mockProducer := new(queue.MockEmailProducer)
	mockAudit := new(MockAuditClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := &config.AppConfig{
		EnableLowStockAlert: true,
		LowStockThreshold:   5,
		LowStockAlertEmail:  "admin@test.com",
	}

	svc := service.NewItemService(mockRepo, mockProducer, mockAudit, logger, cfg)
	return svc, mockRepo, mockProducer, mockAudit
}

// --- TEST CASES ---

func TestItemService_GetByID(t *testing.T) {
	svc, mockRepo, _, _ := setupItemService(t)

	t.Run("Found", func(t *testing.T) {
		expectedItem := &entity.Item{ID: 1, Name: "Kopi", Stock: 10}

		// Mock behavior
		mockRepo.On("FindByID", 1).Return(expectedItem, nil).Once()

		item, err := svc.GetByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, "Kopi", item.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		// Mock behavior: return nil object, nil error (atau error not found tergantung implementasi repo)
		mockRepo.On("FindByID", 99).Return(nil, nil).Once()

		item, err := svc.GetByID(context.Background(), 99)

		assert.NoError(t, err)
		assert.Nil(t, item)
		mockRepo.AssertExpectations(t)
	})
}

func TestItemService_GetAll(t *testing.T) {
	svc, mockRepo, _, _ := setupItemService(t)

	t.Run("Success", func(t *testing.T) {
		mockItems := []entity.Item{
			{ID: 1, Name: "A"}, {ID: 2, Name: "B"},
		}
		// CountAll harus dipanggil dulu karena logic di service memanggil CountAll
		mockRepo.On("CountAll", "", 0).Return(int64(2), nil).Once()
		mockRepo.On("FindAll", 10, 0, "", 0).Return(mockItems, nil).Once()

		items, total, err := svc.GetAll(context.Background(), 10, 0, "", 0)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, items, 2)
	})

	t.Run("Empty Data", func(t *testing.T) {
		mockRepo.On("CountAll", "", 0).Return(int64(0), nil).Once()
		// FindAll tidak akan dipanggil jika CountAll 0

		items, total, err := svc.GetAll(context.Background(), 10, 0, "", 0)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Nil(t, items)
	})
}

func TestItemService_Create(t *testing.T) {
	svc, mockRepo, _, _ := setupItemService(t)

	newItem := entity.Item{Name: "New Item", SKU: "NEW-001"}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Create", newItem).Return(nil).Once()

		err := svc.Create(context.Background(), newItem)
		assert.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		mockRepo.On("Create", newItem).Return(errors.New("db error")).Once()

		err := svc.Create(context.Background(), newItem)
		assert.Error(t, err)
	})
}

func TestItemService_Update(t *testing.T) {
	svc, mockRepo, _, _ := setupItemService(t)

	id := 1
	updateReq := entity.UpdateItemRequest{Name: "Updated Name"}

	mockRepo.On("Update", id, updateReq).Return(nil).Once()

	err := svc.Update(context.Background(), id, updateReq)
	assert.NoError(t, err)
}

func TestItemService_Delete(t *testing.T) {
	svc, mockRepo, _, _ := setupItemService(t)

	mockRepo.On("Delete", 1).Return(nil).Once()

	err := svc.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestItemService_UpdateStock_Logic(t *testing.T) {
	// Setup
	svc, mockRepo, mockProducer, mockAudit := setupItemService(t)

	// Table Driven Test untuk berbagai skenario stok
	tests := []struct {
		name        string
		oldStock    int
		newStock    int
		reqQty      int // Qty perubahan (hanya untuk logik mock)
		expectEmail bool
	}{
		{
			name:        "Stay Above Threshold (10 to 6)",
			oldStock:    10,
			newStock:    6,
			reqQty:      4,
			expectEmail: false,
		},
		{
			name:        "Cross Threshold (10 to 3)",
			oldStock:    10,
			newStock:    3,
			reqQty:      7,
			expectEmail: true,
		},
		{
			name:        "Already Below Threshold (4 to 2)",
			oldStock:    4,
			newStock:    2,
			reqQty:      2,
			expectEmail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemID := 100
			user := &entity.User{ID: 1, Username: "Budi", Role: "Staff"}
			req := entity.UpdateStockRequest{ItemID: itemID, Type: "OUT", Quantity: tt.reqQty}

			// Data Sebelum Update
			itemBefore := &entity.Item{ID: itemID, Name: "Test Item", SKU: "TEST-001", Stock: tt.oldStock}
			// Data Sesudah Update
			itemAfter := &entity.Item{ID: itemID, Name: "Test Item", Stock: tt.newStock}

			// 1. Mock FindByID (Sebelum)
			mockRepo.On("FindByID", itemID).Return(itemBefore, nil).Once()

			// 2. Mock UpdateStock (Action DB)
			mockRepo.On("UpdateStock", req, user.ID).Return(nil).Once()

			// 3. Mock FindByID (Sesudah, untuk cek logic alert)
			// Method ini hanya dipanggil jika oldStock ditemukan
			mockRepo.On("FindByID", itemID).Return(itemAfter, nil).Once()

			// 4. Mock Producer (Hanya jika expectEmail = true)
			if tt.expectEmail {
				mockProducer.On("PushEmailJob", mock.Anything, mock.MatchedBy(func(job entity.EmailJob) bool {
					return job.Subject == "LOW STOCK: Test Item"
				})).Return(nil).Once()
			}

			// 5. Mock Audit (Expect LogActivity called)
			mockAudit.On("LogActivity", mock.Anything).Return().Once()

			// EXECUTE
			err := svc.UpdateStock(context.Background(), req, user)

			// ASSERT
			assert.NoError(t, err)

			// Tunggu sebentar karena Email dikirim via Goroutine
			if tt.expectEmail {
				time.Sleep(500 * time.Millisecond)
			}

			mockRepo.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
			mockAudit.AssertExpectations(t)
		})
	}
}
