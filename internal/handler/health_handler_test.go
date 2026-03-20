package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"warehouse-management-api/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Check(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Success - Service Healthy", func(t *testing.T) {
		mockRepo := new(repository.MockHealthRepo)
		mockRepo.On("Ping").Return(nil) // Mock sukses

		handler := NewHealthHandler(mockRepo, logger)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.Check(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - DB Connection Error", func(t *testing.T) {
		mockRepo := new(repository.MockHealthRepo)
		mockRepo.On("Ping").Return(errors.New("db connection lost")) // Mock error

		handler := NewHealthHandler(mockRepo, logger)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.Check(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
