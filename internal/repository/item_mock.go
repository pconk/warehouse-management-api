package repository

import (
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockItemRepo struct {
	mock.Mock
}

func (m *MockItemRepo) FindByID(id int) (*entity.Item, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}
func (m *MockItemRepo) CountAll(name string, categoryID int) (int64, error) {
	args := m.Called(name, categoryID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockItemRepo) FindAll(limit, offset int, name string, categoryID int) ([]entity.Item, error) {
	args := m.Called(limit, offset, name, categoryID)
	return args.Get(0).([]entity.Item), args.Error(1)
}

func (m *MockItemRepo) Create(item entity.Item) error {
	return m.Called(item).Error(0)
}

func (m *MockItemRepo) Update(id int, item entity.UpdateItemRequest) error {
	return m.Called(id, item).Error(0)
}

func (m *MockItemRepo) UpdateStock(req entity.UpdateStockRequest, userID int64) error {
	return m.Called(req, userID).Error(0)
}

func (m *MockItemRepo) Delete(id int) error {
	return m.Called(id).Error(0)
}

func (m *MockItemRepo) GetStockLogs(limit, offset int, itemID int, logType string) ([]entity.StockLog, int64, error) {
	args := m.Called(limit, offset, itemID, logType)
	var total int64
	if args.Get(1) != nil {
		total = args.Get(1).(int64)
	}

	return args.Get(0).([]entity.StockLog), total, args.Error(2)
}
