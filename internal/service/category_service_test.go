package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/repository"
	"warehouse-management-api/internal/service"

	"github.com/stretchr/testify/assert"
)

// Helper setup
func setupCategoryService() (service.CategoryServiceInterface, *repository.MockCategoryRepo) {
	mockRepo := new(repository.MockCategoryRepo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := service.NewCategoryService(mockRepo, logger)
	return svc, mockRepo
}

func TestCategoryService_GetAll(t *testing.T) {
	svc, mockRepo := setupCategoryService()

	t.Run("Success", func(t *testing.T) {
		mockData := []entity.Category{
			{ID: 1, Name: "A"}, {ID: 2, Name: "B"},
		}
		mockRepo.On("FindAll").Return(mockData, nil).Once()

		result, err := svc.GetAll(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		// Simulasi DB Error
		mockRepo.On("FindAll").Return([]entity.Category{}, errors.New("db connection failed")).Once()

		result, err := svc.GetAll(context.Background())

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "db connection failed", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_GetByID(t *testing.T) {
	svc, mockRepo := setupCategoryService()

	t.Run("Found", func(t *testing.T) {
		mockCat := &entity.Category{ID: 1, Name: "Elektronik"}
		mockRepo.On("FindByID", 1).Return(mockCat, nil).Once()

		result, err := svc.GetByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, "Elektronik", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", 99).Return(nil, errors.New("not found")).Once()

		result, err := svc.GetByID(context.Background(), 99)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_Create(t *testing.T) {
	svc, mockRepo := setupCategoryService()

	t.Run("Success", func(t *testing.T) {
		newCat := &entity.Category{Name: "Elektronik"}
		mockRepo.On("Create", newCat).Return(nil).Once()

		err := svc.Create(context.Background(), newCat)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		newCat := &entity.Category{Name: "ErrorCat"}
		mockRepo.On("Create", newCat).Return(errors.New("create failed")).Once()

		err := svc.Create(context.Background(), newCat)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_Update(t *testing.T) {
	svc, mockRepo := setupCategoryService()

	cat := &entity.Category{ID: 1, Name: "Updated"}
	mockRepo.On("Update", cat).Return(nil).Once()

	err := svc.Update(context.Background(), cat)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete(t *testing.T) {
	svc, mockRepo := setupCategoryService()

	mockRepo.On("Delete", 1).Return(nil).Once()

	err := svc.Delete(context.Background(), 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
