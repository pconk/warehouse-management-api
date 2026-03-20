package repository

import (
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockCategoryRepo struct {
	mock.Mock
}

func (m *MockCategoryRepo) FindAll() ([]entity.Category, error) {
	args := m.Called()
	return args.Get(0).([]entity.Category), args.Error(1)
}
func (m *MockCategoryRepo) FindByID(id int) (*entity.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}
func (m *MockCategoryRepo) Create(category *entity.Category) error {
	return m.Called(category).Error(0)
}
func (m *MockCategoryRepo) Update(category *entity.Category) error {
	return m.Called(category).Error(0)
}
func (m *MockCategoryRepo) Delete(id int) error {
	return m.Called(id).Error(0)
}
