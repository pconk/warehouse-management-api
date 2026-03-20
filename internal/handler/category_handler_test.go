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
	"warehouse-management-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- TESTS ---

func TestCategoryHandler_GetAll_Success(t *testing.T) {
	mockService := new(service.MockCategoryService)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Service: mockService, Logger: logger}

	mockData := []entity.Category{
		{ID: 1, Name: "Elektronik"},
		{ID: 2, Name: "Furniture"},
	}

	// Expectation: Service.GetAll dipanggil
	mockService.On("GetAll", mock.Anything).Return(mockData, nil)

	req := httptest.NewRequest("GET", "/categories", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_GetByID_Found(t *testing.T) {
	mockService := new(service.MockCategoryService)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Service: mockService, Logger: logger}

	mockCat := &entity.Category{ID: 1, Name: "Elektronik"}

	mockService.On("GetByID", mock.Anything, 1).Return(mockCat, nil)

	req := httptest.NewRequest("GET", "/categories/1", nil)
	// Setup URL Param Chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_GetByID_NotFound(t *testing.T) {
	mockService := new(service.MockCategoryService)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Service: mockService, Logger: logger}

	// Service return nil jika tidak ketemu (atau error not found tergantung implementasi service)
	mockService.On("GetByID", mock.Anything, 99).Return(nil, errors.New("category not found"))

	req := httptest.NewRequest("GET", "/categories/99", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryHandler_Create_Success(t *testing.T) {
	mockService := new(service.MockCategoryService)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Service: mockService, Logger: logger}

	// Service return nil error -> sukses
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*entity.Category")).Return(nil)

	body := `{"name": "Otomotif"}`
	req := httptest.NewRequest("POST", "/categories", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	mockService := new(service.MockCategoryService)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Service: mockService, Logger: logger}

	mockService.On("Update", mock.Anything, mock.MatchedBy(func(cat *entity.Category) bool {
		return cat.ID == 1 && cat.Name == "Otomotif Updated"
	})).Return(nil)

	body := `{"name": "Otomotif Updated"}`
	req := httptest.NewRequest("PUT", "/categories/1", strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	mockService := new(service.MockCategoryService)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Service: mockService, Logger: logger}

	mockService.On("Delete", mock.Anything, 1).Return(nil)

	req := httptest.NewRequest("DELETE", "/categories/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_InvalidJSON(t *testing.T) {
	h := CategoryHandler{Logger: slog.Default()}

	req := httptest.NewRequest("POST", "/categories", strings.NewReader(`{invalid-json}`))
	w := httptest.NewRecorder()

	h.Create(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
