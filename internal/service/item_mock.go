package service

import (
	"context"
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockItemService struct {
	mock.Mock
}

func (m *MockItemService) UpdateStock(ctx context.Context, req entity.UpdateStockRequest, userID int) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}

// Tambahkan method lain dari interface jika ada (misal GetAll, GetByID)
