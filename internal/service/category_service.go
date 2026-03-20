package service

import (
	"context"
	"log/slog"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/repository"
)

type CategoryServiceInterface interface {
	GetAll(ctx context.Context) ([]entity.Category, error)
	GetByID(ctx context.Context, id int) (*entity.Category, error)
	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id int) error
}

type categoryService struct {
	repo   repository.CategoryRepositoryInterface
	logger *slog.Logger
}

func NewCategoryService(repo repository.CategoryRepositoryInterface, logger *slog.Logger) CategoryServiceInterface {
	return &categoryService{
		repo:   repo,
		logger: logger,
	}
}

func (s *categoryService) GetAll(ctx context.Context) ([]entity.Category, error) {
	return s.repo.FindAll()
}

func (s *categoryService) GetByID(ctx context.Context, id int) (*entity.Category, error) {
	return s.repo.FindByID(id)
}

func (s *categoryService) Create(ctx context.Context, category *entity.Category) error {
	return s.repo.Create(category)
}

func (s *categoryService) Update(ctx context.Context, category *entity.Category) error {
	// Bisa tambahkan validasi bisnis di sini jika perlu
	return s.repo.Update(category)
}

func (s *categoryService) Delete(ctx context.Context, id int) error {
	// Bisa tambahkan cek apakah kategori masih dipakai oleh item lain sebelum delete
	return s.repo.Delete(id)
}
