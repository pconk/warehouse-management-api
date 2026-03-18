package handler

import (
	"log/slog"
	"math"
	"net/http"
	"warehouse-management-api/internal/helper"
	"warehouse-management-api/internal/repository"
)

type CategoryHandler struct {
	Repo   repository.CategoryRepositoryInterface // Ganti dari *repository.CategoryRepository
	Logger *slog.Logger
}

func NewCategoryHandler(repo repository.CategoryRepositoryInterface, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{
		Repo:   repo,
		Logger: logger,
	}
}

// GetAll godoc
// @Summary Ambil semua kategori
// @Description Mengambil daftar kategori barang untuk keperluan dropdown atau filter
// @Tags categories
// @Produce json
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Data per halaman" default(10)
// @Success 200 {object} helper.WebResponsePaging{data=[]entity.Category}
// @Router /categories [get]
// @Security ApiKeyAuth
func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, offset, page := helper.GetPaginationParams(r)

	categories, total, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		h.Logger.Error("Failed to fetch categories", "error", err.Error())
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Internal server error", nil)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := &helper.PaginationMeta{
		CurrentPage: page,
		TotalItems:  total,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	helper.SendResponseWithPaging(w, http.StatusOK, "OK", "Categories retrieved", categories, meta)
}
