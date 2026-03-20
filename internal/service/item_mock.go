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

func (m *MockItemService) GetAll(ctx context.Context, limit, offset int, filterName string, filterCatID int) ([]entity.Item, int64, error) {
	args := m.Called(ctx, limit, offset, filterName, filterCatID)
	return args.Get(0).([]entity.Item), args.Get(1).(int64), args.Error(2)
}
func (m *MockItemService) GetAllForExport(ctx context.Context) ([]entity.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.Item), args.Error(1)
}
func (m *MockItemService) GetByID(ctx context.Context, id int) (*entity.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}
func (m *MockItemService) Create(ctx context.Context, item entity.Item) error {
	return m.Called(ctx, item).Error(0)
}
func (m *MockItemService) Update(ctx context.Context, id int, item entity.UpdateItemRequest) error {
	return m.Called(ctx, id, item).Error(0)
}
func (m *MockItemService) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockItemService) GetStockLogs(ctx context.Context, limit, offset int, itemID int, logType string) ([]entity.StockLog, int64, error) {
	args := m.Called(ctx, limit, offset, itemID, logType)
	return args.Get(0).([]entity.StockLog), args.Get(1).(int64), args.Error(2)
}
