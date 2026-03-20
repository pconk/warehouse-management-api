package service

import (
	"context"
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) GetAll(ctx context.Context) ([]entity.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.Category), args.Error(1)
}

func (m *MockCategoryService) GetByID(ctx context.Context, id int) (*entity.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryService) Create(ctx context.Context, category *entity.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *MockCategoryService) Update(ctx context.Context, category *entity.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *MockCategoryService) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
