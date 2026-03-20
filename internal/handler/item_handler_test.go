package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"warehouse-management-api/internal/entity"
	"warehouse-management-api/internal/middleware"
	"warehouse-management-api/internal/repository"
	"warehouse-management-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByID_Success(t *testing.T) {
	// 1. Setup Mock & Handler
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// 2. Ekspektasi: Jika ID 1 dipanggil, kembalikan data item
	itemData := &entity.Item{ID: 1, Name: "Test Item", SKU: "Tst-01"}
	mockRepo.On("FindByID", 1).Return(itemData, nil)

	// 3. Buat Request ke endpoint /items/1
	req := httptest.NewRequest("GET", "/items/1", nil)

	// Tambahkan URL Param Chi secara manual untuk testing
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// 4. Jalankan Handler
	h.GetByID(w, req)

	// 5. Assert (Cek hasil)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "OK", response["status"])
}

func TestGetByID_NotFound(t *testing.T) {
	// 1. Setup Mock & Handler
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	mockRepo.On("FindByID", 1).Return(nil, nil)

	// 3. Buat Request ke endpoint /items/1
	req := httptest.NewRequest("GET", "/items/1", nil)

	// Tambahkan URL Param Chi secara manual untuk testing
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// 4. Jalankan Handler
	h.GetByID(w, req)

	// 5. Assert (Cek hasil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAllItem_Success(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	items := []entity.Item{
		{ID: 1, SKU: "SKU01", Name: "Item 1", Price: 1000, Stock: 10, CategoryName: "Elektronik"},
		{ID: 2, SKU: "SKU02", Name: "Item 2", Price: 2000, Stock: 5, CategoryName: "Alat Kantor"},
	}

	// Simulasi DB mati/koneksi putus
	mockRepo.On("List", 10, 0, "", 0).
		Return(items, int64(2), nil)

	req := httptest.NewRequest("GET", "/items?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	h.GetAllItem(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "OK", response["status"])
	assert.NotNil(t, response["pagination"])

	// Cek apakah data muncul
	data := response["data"].([]interface{})
	assert.True(t, len(data) > 0)

	mockRepo.AssertExpectations(t)
}

func TestGetAllItem_DatabaseError(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// Simulasi DB mati/koneksi putus
	mockRepo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]entity.Item{}, int64(0), errors.New("connection refused"))

	req := httptest.NewRequest("GET", "/items", nil)
	w := httptest.NewRecorder()

	h.GetAllItem(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Internal Server Error", response["status"])
}

func TestCreateItem_Success(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger, Validate: validate}

	newItem := entity.Item{
		CategoryID: 1,
		SKU:        "MAC-001",
		Name:       "Macbook Pro",
		Price:      20000000,
		Stock:      10,
	}

	mockRepo.On("Insert", mock.AnythingOfType("entity.Item")).Return(nil)

	body, _ := json.Marshal(newItem)
	req := httptest.NewRequest("POST", "/items/create", strings.NewReader(string(body)))
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateItem_DuplicateSKU(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger, Validate: validate}

	mockRepo.On("Insert", mock.Anything).Return(errors.New("Duplicate entry 'MAC-001'"))

	body := `{"category_id": 1, "sku": "MAC-001", "name": "Mac", "price": 1000, "stock": 1}`
	req := httptest.NewRequest("POST", "/items/create", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusConflict, w.Code) // Karena duplikat kita set 400 atau 409
}

func TestUpdateItem_ValidationError(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger, Validate: validate}

	// Body JSON dengan harga minus (Price: -100)
	body := `{"category_id": 1, "name": "Salah Harga", "price": -100}`
	req := httptest.NewRequest("PUT", "/items/1", strings.NewReader(body))

	// Setup Chi Context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Jalankan Handler
	h.Update(w, req)

	// ASSERTION
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Validation Error", response["status"])

	// Pastikan Repo.Update TIDAK PERNAH dipanggil karena sudah gagal di validasi
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateItem_Success(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger, Validate: validate}

	// Data input yang valid
	updateData := entity.UpdateItemRequest{
		CategoryID: 1,
		Name:       "Monitor Gaming LG",
		Price:      3500000,
	}

	// Ekspektasi: Repo.Update dipanggil dengan ID 1 dan data di atas, kembalikan nil (no error)
	mockRepo.On("Update", 1, updateData).Return(nil)

	body, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", "/items/1", strings.NewReader(string(body)))

	// Setup Context Chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t) // Pastikan repo benar-benar dipanggil
}

func TestUpdateStock_Table(t *testing.T) {
	// Setup dasar
	mockService := new(service.MockItemService) // Sekarang nge-mock Service, bukan Repo
	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Handler sekarang menggunakan Service
	h := ItemHandler{
		Service:  mockService,
		Logger:   logger,
		Validate: validate,
	}

	user := &entity.User{ID: 1, Username: "admin"}

	tests := []struct {
		name           string
		inputBody      string
		mockReturn     error
		expectedStatus int
	}{
		{
			name:           "Success",
			inputBody:      `{"item_id": 1, "type": "IN", "quantity": 10, "reason": "Restock"}`,
			mockReturn:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Validation Error (Missing ItemID)",
			inputBody:      `{"type": "IN", "quantity": 10}`,
			mockReturn:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service Error (Insufficient Stock)",
			inputBody:      `{"item_id": 1, "type": "OUT", "quantity": 100}`,
			mockReturn:     errors.New("insufficient stock"), // Pesan error dari service
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Internal Server Error",
			inputBody:      `{"item_id": 1, "type": "IN", "quantity": 10}`,
			mockReturn:     errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Mock sesuai skenario
			// Kita hanya ekspektasi Service dipanggil jika validasi input di Handler lolos
			if tt.name == "Success" || tt.expectedStatus == http.StatusInternalServerError || (tt.expectedStatus == http.StatusBadRequest && tt.mockReturn != nil) {
				mockService.On("UpdateStock", mock.Anything, mock.Anything, 1).Return(tt.mockReturn).Once()
			}

			req := httptest.NewRequest("POST", "/items/update-stock", strings.NewReader(tt.inputBody))
			ctx := context.WithValue(req.Context(), middleware.UserKey, user)
			w := httptest.NewRecorder()

			h.UpdateStock(w, req.WithContext(ctx))

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Bersihkan mock untuk iterasi selanjutnya
			mockService.ExpectedCalls = nil
		})
	}
}

func TestDeleteItem_NotFound(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// Ekspektasi: Repo.Delete balikin error "item not found"
	mockRepo.On("Delete", 99).Return(errors.New("item not found"))

	req := httptest.NewRequest("DELETE", "/items/99", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Barang tidak ditemukan", response["message"])
}

func TestExportCSV_Success(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// 1. Data dummy untuk di-export
	items := []entity.Item{
		{ID: 1, SKU: "SKU01", Name: "Item 1", Price: 1000, Stock: 10, CategoryName: "Elektronik"},
		{ID: 2, SKU: "SKU02", Name: "Item 2", Price: 2000, Stock: 5, CategoryName: "Alat Kantor"},
	}

	// 2. Setup Mock: FindAllForExport dipanggil satu kali
	mockRepo.On("FindAllForExport").Return(items, nil)

	req := httptest.NewRequest("GET", "/items/export", nil)
	w := httptest.NewRecorder()

	// 3. Jalankan Handler
	h.ExportCSV(w, req)

	// 4. ASSERTIONS
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment;filename=items_export.csv")

	// 5. Cek isi body CSV
	body := w.Body.String()
	assert.Contains(t, body, "ID,SKU,Name,Price,Stock,Category") // Header
	assert.Contains(t, body, "1,SKU01,Item 1,1000.00,10,Elektronik")
	assert.Contains(t, body, "2,SKU02,Item 2,2000.00,5,Alat Kantor")

	mockRepo.AssertExpectations(t)
}

func TestExportCSV_Error(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// Simulasi DB Error saat export
	mockRepo.On("FindAllForExport").Return([]entity.Item{}, errors.New("db connection lost"))

	req := httptest.NewRequest("GET", "/items/export", nil)
	w := httptest.NewRecorder()

	h.ExportCSV(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetStockLogs_Success(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// 1. Data dummy logs
	adminName := "admin"
	logs := []entity.StockLog{
		{ID: 1, ItemID: 1, ItemName: "Macbook", Type: "IN", Quantity: 10, Reason: "Restock", UserName: &adminName},
	}
	total := int64(1)

	// 2. Setup Mock: Pastikan parameter sesuai dengan handler (limit 10, offset 0, itemID 0, type empty)
	mockRepo.On("GetStockLogs", 10, 0, 0, "").Return(logs, total, nil)

	req := httptest.NewRequest("GET", "/items/stock-logs?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	// 3. Jalankan Handler
	h.GetStockLogs(w, req)

	// 4. Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "OK", response["status"])
	assert.NotNil(t, response["pagination"])

	// Cek apakah data log muncul
	data := response["data"].([]interface{})
	assert.True(t, len(data) > 0)

	mockRepo.AssertExpectations(t)
}

func TestGetStockLogs_DatabaseError(t *testing.T) {
	mockRepo := new(repository.MockItemRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := ItemHandler{Repo: mockRepo, Logger: logger}

	// Simulasi DB Error
	mockRepo.On("GetStockLogs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]entity.StockLog{}, int64(0), errors.New("db error"))

	req := httptest.NewRequest("GET", "/items/stock-logs", nil)
	w := httptest.NewRecorder()

	h.GetStockLogs(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
