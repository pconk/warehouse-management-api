package handler

import (
	"log/slog"
	"net/http"
	"warehouse-management-api/internal/helper"
	"warehouse-management-api/internal/repository" // import repo Anda
)

type HealthHandler struct {
	Repo   repository.HealthRepositoryInterface
	Logger *slog.Logger
}

func NewHealthHandler(repo repository.HealthRepositoryInterface, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		Repo:   repo,
		Logger: logger,
	}
}

// Check godoc
// @Summary Health Check
// @Description Mengecek status aplikasi dan koneksi database
// @Tags health
// @Produce json
// @Success 200 {object} helper.WebResponse
// @Router /health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	err := h.Repo.Ping()

	if err != nil {
		h.Logger.Error("Health check failed", "error", err.Error())
		helper.SendResponse(w, http.StatusServiceUnavailable, "Error", "Service is unhealthy", nil)
		return
	}

	helper.SendResponse(w, http.StatusOK, "OK", "Service is healthy", nil)
}
