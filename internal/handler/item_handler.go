package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/helper"
	"warehouse-management-api/internal/middleware"
	"warehouse-management-api/internal/service"

	"encoding/csv"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type ItemHandler struct {
	Service  service.ItemServiceInterface
	Logger   *slog.Logger
	Validate *validator.Validate
}

func NewItemHandler(service service.ItemServiceInterface, logger *slog.Logger) *ItemHandler {
	return &ItemHandler{
		Service:  service,
		Logger:   logger,
		Validate: validator.New(),
	}
}

// GetByID godoc
// @Summary Ambil barang detail
// @Description Mengambil detail satu barang berdasarkan ID
// @Tags items
// @Produce json
// @Param id path int true "ID Barang"
// @Success 200 {object} helper.WebResponse{data=entity.Item}
// @Router /items/{id} [get]
// @Security BearerAuth
func (h *ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id == 0 {
		helper.SendResponse(w, http.StatusBadRequest, "Fail", "ID harus angka", nil)
		return
	}

	item, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		h.Logger.Error("GetByID  failed", "error", err.Error(), "request_id", middleware.GetRequestID(r.Context()))

		// Response ke user (aman, tidak bocor detail DB)
		helper.SendResponse(w, http.StatusInternalServerError, "Internal Server Error", "Gagal mengambil data dari server", nil)
		return
	}

	if item == nil {
		helper.SendResponse(w, http.StatusNotFound, "Fail", "Barang tidak ditemukan", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "OK", "Data barang berhasil diambil", item)
}

// GetAllItem godoc
// @Summary Ambil semua barang
// @Description Mengambil daftar barang dengan fitur paging dan filter
// @Tags items
// @Accept json
// @Produce json
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Data per halaman" default(10)
// @Param name query string false "Filter nama"
// @Param category_id query int false "Filter kategori ID"
// @Success 200 {object} helper.WebResponsePaging{data=[]entity.Item}
// @Router /items [get]
// @Security BearerAuth
func (h *ItemHandler) GetAllItem(w http.ResponseWriter, r *http.Request) {

	limit, offset, page := helper.GetPaginationParams(r)

	filterName := r.URL.Query().Get("name")
	filterCatID, _ := strconv.Atoi(r.URL.Query().Get("category_id"))

	items, totalItems, err := h.Service.GetAll(r.Context(), limit, offset, filterName, filterCatID)

	if err != nil {
		// Log error asli (untuk internal dev)
		h.Logger.Error("Get All Item failed", "error", err.Error(), "request_id", middleware.GetRequestID(r.Context()))

		helper.SendResponse(w, http.StatusInternalServerError, "Internal Server Error", "Gagal mengambil data dari server", nil)
		return
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	pagination := helper.PaginationMeta{
		CurrentPage: page,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	// 3. Pastikan jika data kosong, balikan array [] bukan null
	if items == nil {
		items = []entity.Item{}
	}

	// 4. Response Sukses
	helper.SendResponseWithPaging(w, http.StatusOK, "OK", "Data barang berhasil diambil", items, &pagination)
}

// Create godoc
// @Summary Tambah barang baru
// @Description Contoh data JSON:
// @Description `{"category_id":1,"sku":"ELC-MON-003","name":"Dell UltraSharp 27-inch","price":8500000.00,"stock":5}`
// @Description
// @Description **Note:** Pastikan SKU belum terdaftar di sistem.
// @Tags items
// @Accept json
// @Produce json
// @Param item body entity.Item true "Data barang"
// @Success 201 {object} helper.WebResponse
// @Failure 400 {object} helper.WebResponse
// @Router /items/create [post]
// @Security BearerAuth
func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var newItem entity.Item

	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		helper.SendResponse(w, http.StatusBadRequest, "Bad Request", "Invalid JSON", nil)
		return
	}

	if err := h.Validate.Struct(newItem); err != nil {
		// Panggil helper untuk merapikan error
		prettyErrors := helper.FormatValidationError(err)

		helper.SendResponse(w, http.StatusBadRequest, "Validation Error", "Beberapa field tidak valid", prettyErrors)
		return
	}

	if err := h.Service.Create(r.Context(), newItem); err != nil {
		// 1. Log error asli untuk debugging di server
		h.Logger.Error("Create Item failed", "error", err.Error(), "request_id", middleware.GetRequestID(r.Context()))

		// 2. Cek apakah ini error duplikat (misal SKU sudah terdaftar)
		if strings.Contains(err.Error(), "Duplicate entry") {
			helper.SendResponse(w, http.StatusConflict, "Fail", "Item with this SKU already exists", nil)
			return
		}

		// 3. Untuk error teknis lainnya (Koneksi mati, dsb), berikan status 500
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "We are experiencing technical difficulties, please try again later", nil)
		return
	}

	// 4. Kasih response sukses
	helper.SendResponse(w, http.StatusCreated, "OK", "Item created successfully", nil)

}

// Update godoc
// @Summary Update metadata barang
// @Description Mengubah Nama, Harga, atau Kategori (SKU & Stok terkunci)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID Barang"
// @Param item body entity.UpdateItemRequest true "Data update"
// @Success 200 {object} helper.WebResponse
// @Router /items/{id} [put]
// @Security BearerAuth
func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id == 0 {
		helper.SendResponse(w, http.StatusBadRequest, "Fail", "ID harus angka", nil)
		return
	}

	var updateReq entity.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		// h.Logger.Error("error", "id", id, "updateReq", updateReq)
		helper.SendResponse(w, http.StatusBadRequest, "Fail", "Format JSON tidak valid", nil)
		return
	}

	// Validasi input (Name, Price, CategoryID wajib ada)
	if err := h.Validate.Struct(updateReq); err != nil {
		helper.SendResponse(w, http.StatusBadRequest, "Validation Error", "Data tidak valid", helper.FormatValidationError(err))
		return
	}

	if err := h.Service.Update(r.Context(), id, updateReq); err != nil {
		h.Logger.Error("Update failed", "error", err.Error(), "request_id", middleware.GetRequestID(r.Context()))

		if err.Error() == "item not found or no changes made" {
			helper.SendResponse(w, http.StatusNotFound, "Fail", "Barang tidak ditemukan atau tidak ada perubahan", nil)
			return
		}

		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Gagal memperbarui data", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "OK", "Data barang berhasil diperbarui", nil)
}

// UpdateStock godoc
// @Summary Update stok barang
// @Description Menambah (IN) atau mengurangi (OUT) stok dengan audit log
// @Tags admin
// @Accept json
// @Produce json
// @Param req body entity.UpdateStockRequest true "Request update stok"
// @Success 200 {object} helper.WebResponse
// @Router /items/update-stock [post]
// @Security BearerAuth
func (h *ItemHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var updateRequest entity.UpdateStockRequest

	// 2. Decode JSON dari Request Body ke Struct
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		helper.SendResponse(w, http.StatusBadRequest, "Bad Request", "Invalid JSON format", nil)
		return
	}

	// 2. Validasi (Gunakan validator)
	if err := h.Validate.Struct(updateRequest); err != nil {
		// Panggil helper untuk merapikan error
		prettyErrors := helper.FormatValidationError(err)

		// Kirim response dengan status 400 tapi data berisi detail errornya
		helper.SendResponse(w, http.StatusBadRequest, "Validation Error", "Beberapa field tidak valid", prettyErrors)
		return
	}

	// Ambil user menggunakan helper function yang sudah kita buat
	user := middleware.GetUser(r.Context())

	// Panggil Service (Logic ada di dalam sini)
	err := h.Service.UpdateStock(r.Context(), updateRequest, user)
	if err != nil {
		h.Logger.Error("Update stock failed", "error", err.Error(), "request_id", middleware.GetRequestID(r.Context()))

		if strings.Contains(err.Error(), "insufficient") {
			helper.SendResponse(w, http.StatusBadRequest, "Fail", err.Error(), nil)
			return
		}

		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Internal Server Error", nil)
		return
	}
	helper.SendResponse(w, http.StatusOK, "OK", "Stock updated successfully", nil)

}

// Delete godoc
// @Summary Soft delete barang
// @Description Menghapus barang secara logis (data tetap di DB)
// @Tags admin
// @Param id path int true "ID Barang"
// @Success 200 {object} helper.WebResponse
// @Router /items/{id} [delete]
// @Security BearerAuth
func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id == 0 {
		helper.SendResponse(w, http.StatusBadRequest, "Fail", "ID harus angka", nil)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		if err.Error() == "item not found" {
			helper.SendResponse(w, http.StatusNotFound, "Fail", "Barang tidak ditemukan", nil)
			return
		}
		h.Logger.Error("Delete failed", "error", err.Error(),
			"request_id", middleware.GetRequestID(r.Context()))
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Gagal menghapus data", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "OK", "Barang berhasil dihapus (Soft Delete)", nil)
}

// ExportCSV godoc
// @Summary Export data barang ke CSV
// @Description Mengunduh seluruh daftar barang (yang belum dihapus) dalam format file .csv
// @Tags items
// @Produce text/csv
// @Success 200 {file} file "items_export.csv"
// @Failure 500 {object} helper.WebResponse
// @Router /items/export [get]
// @Security BearerAuth
func (h *ItemHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.GetAllForExport(r.Context())
	if err != nil {
		h.Logger.Error("Delete failed", "error", err.Error(),
			"request_id", middleware.GetRequestID(r.Context()))
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Gagal export data", nil)
		return
	}

	// Set header agar browser mendownload sebagai file .csv
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=items_export.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Tulis Header CSV
	writer.Write([]string{"ID", "SKU", "Name", "Price", "Stock", "Category"})

	// Tulis Data
	for _, item := range items {
		row := []string{
			strconv.Itoa(item.ID),
			item.SKU,
			item.Name,
			fmt.Sprintf("%.2f", item.Price),
			strconv.Itoa(item.Stock),
			item.CategoryName,
		}
		writer.Write(row)
	}
}

// GetStockLogs godoc
// @Summary Ambil riwayat stok
// @Description Mengambil log aktivitas stok (IN/OUT) dengan filter item dan tipe
// @Tags items
// @Produce json
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Data per halaman" default(10)
// @Param item_id query int false "Filter berdasarkan ID Barang"
// @Param type query string false "Filter tipe log (IN/OUT)" Enums(IN, OUT)
// @Success 200 {object} helper.WebResponsePaging{data=[]entity.StockLog}
// @Router /items/stock-logs [get]
// @Security BearerAuth
func (h *ItemHandler) GetStockLogs(w http.ResponseWriter, r *http.Request) {
	limit, offset, page := helper.GetPaginationParams(r)

	// Filter opsional
	itemID, _ := strconv.Atoi(r.URL.Query().Get("item_id"))
	logType := r.URL.Query().Get("type") // IN atau OUT

	if logType != "" && logType != "IN" && logType != "OUT" {
		helper.SendResponse(w, http.StatusBadRequest, "Error", "type hanya IN atau OUT", nil)
		return
	}

	logs, total, err := h.Service.GetStockLogs(r.Context(), limit, offset, itemID, logType)
	if err != nil {
		h.Logger.Error("GetStockLogs  failed", "error", err.Error(),
			"request_id", middleware.GetRequestID(r.Context()))
		helper.SendResponse(w, http.StatusInternalServerError, "Error", "Gagal mengambil riwayat stok", nil)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	pagination := &helper.PaginationMeta{
		CurrentPage: page,
		TotalItems:  total,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	helper.SendResponseWithPaging(w, http.StatusOK, "OK", "Riwayat stok berhasil diambil", logs, pagination)
}
