package handler

import (
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockCategoryRepo struct {
	mock.Mock
}

func (m *MockCategoryRepo) FindAll(limit, offset int) ([]entity.Category, int64, error) {
	args := m.Called(limit, offset)

	var total int64
	if args.Get(1) != nil {
		total = args.Get(1).(int64)
	}

	return args.Get(0).([]entity.Category), total, args.Error(2)
}
