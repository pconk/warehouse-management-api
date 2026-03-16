package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"warehouse-management-api/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllCategories_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Repo: mockRepo, Logger: logger}

	categories := []entity.Category{{ID: 1, Name: "Elektronik"}}
	mockRepo.On("FindAll", 10, 0).Return(categories, int64(1), nil)

	req := httptest.NewRequest("GET", "/categories?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAllCategories_DatabaseError(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := CategoryHandler{Repo: mockRepo, Logger: logger}

	// Simulasi error database
	mockRepo.On("FindAll", mock.Anything, mock.Anything).
		Return([]entity.Category{}, int64(0), errors.New("db connection lost"))

	req := httptest.NewRequest("GET", "/categories", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Error", response["status"])
}
