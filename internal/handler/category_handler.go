package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/helper"
	"warehouse-management-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	Service service.CategoryServiceInterface
	Logger  *slog.Logger
}

func NewCategoryHandler(service service.CategoryServiceInterface, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{Service: service, Logger: logger}
}

// GetAll godoc
// @Summary Ambil semua kategori
// @Description Mengambil daftar semua kategori
// @Tags categories
// @Produce json
// @Success 200 {object} helper.WebResponse{data=[]entity.Category}
// @Router /categories [get]
// @Security BearerAuth
func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Service.GetAll(r.Context())
	if err != nil {
		h.Logger.Error("Failed to fetch categories", "error", err)
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Internal Server Error", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "Success", "List of categories", categories)
}

// GetByID godoc
// @Summary Ambil detail kategori
// @Description Mengambil data kategori berdasarkan ID
// @Tags categories
// @Produce json
// @Param id path int true "ID Kategori"
// @Success 200 {object} helper.WebResponse{data=entity.Category}
// @Router /categories/{id} [get]
// @Security BearerAuth
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	category, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		helper.SendResponse(w, http.StatusNotFound, "Not Found", "Category not found", nil)
		return
	}
	helper.SendResponse(w, http.StatusOK, "Success", "Category detail", category)
}

// Create godoc
// @Summary Buat kategori baru
// @Description Menambahkan kategori baru (Khusus Admin)
// @Tags admin
// @Accept json
// @Produce json
// @Param category body entity.Category true "Data Kategori"
// @Success 201 {object} helper.WebResponse{data=entity.Category}
// @Router /categories [post]
// @Security BearerAuth
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.SendResponse(w, http.StatusBadRequest, "Bad Request", "Invalid JSON", nil)
		return
	}

	if err := h.Service.Create(r.Context(), &req); err != nil {
		h.Logger.Error("Failed to create category", "error", err)
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Failed to create category", nil)
		return
	}

	helper.SendResponse(w, http.StatusCreated, "Success", "Category created", req)
}

// Update godoc
// @Summary Update kategori
// @Description Mengubah data kategori (Khusus Admin)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID Kategori"
// @Param category body entity.Category true "Data Kategori"
// @Success 200 {object} helper.WebResponse
// @Router /categories/{id} [put]
// @Security BearerAuth
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req entity.Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.SendResponse(w, http.StatusBadRequest, "Bad Request", "Invalid JSON", nil)
		return
	}
	req.ID = id

	if err := h.Service.Update(r.Context(), &req); err != nil {
		h.Logger.Error("Failed to update category", "error", err)
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Failed to update category", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "Success", "Category updated", nil)
}

// Delete godoc
// @Summary Hapus kategori
// @Description Menghapus kategori permanen (Khusus Admin)
// @Tags admin
// @Produce json
// @Param id path int true "ID Kategori"
// @Success 200 {object} helper.WebResponse
// @Router /categories/{id} [delete]
// @Security BearerAuth
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if err := h.Service.Delete(r.Context(), id); err != nil {
		h.Logger.Error("Failed to delete category", "error", err)
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Failed to delete category", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "Success", "Category deleted", nil)
}
