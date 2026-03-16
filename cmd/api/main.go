package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "warehouse-management-api/docs"
	"warehouse-management-api/internal/config"
	"warehouse-management-api/internal/handler"
	"warehouse-management-api/internal/middleware"
	"warehouse-management-api/internal/repository"
	"warehouse-management-api/pkg/database"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Warehouse Management API
// @version 1.0
// @description API untuk manajemen stok gudang.
// @description Testing Keys:
// @description - Admin: `secret-admin-key` (Full Access)
// @description - Staff: `secret-staff-key` (Read & Update Only)
// @tag.name admin
// @tag.description Endpoint internal yang memerlukan hak akses Administrator
// @tag.name categories
// @tag.description Endpoint kategori barang
// @tag.name items
// @tag.description Endpoint berhubungan dengan barang dan inventaris
// @tag.name health
// @tag.description Endpoint untuk pengecekan koneksi database

// @termsOfService http://swagger.io

// @contact.name API Support
// @contact.url http://www.swagger.io
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @description Masukkan API Key kamu. Format: [your-api-key]
func main() {
	// Setup JSON Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("configuration file not found")
		os.Exit(1)
	}

	// Koneksi DB
	db, err := database.ConnectDB(cfg.Db)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Init Repository
	healthRepo := repository.NewHealthRepository(db)
	itemRepo := repository.NewItemRepository(db)
	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCategoryRepository(db)

	// Init Handler (Sambil lempar repo-nya)
	healthHandler := handler.NewHealthHandler(healthRepo, logger)
	itemHandler := handler.NewItemHandler(itemRepo, logger)
	catHandler := handler.NewCategoryHandler(catRepo, logger)

	// Inisialisasi Middleware dengan API Key dari .env
	auth := middleware.AuthMiddleware(userRepo, logger)
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 1. Setup Router Chi
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)

	// Middleware global (untuk semua route)
	r.Use(middleware.Logger(logger))

	r.Get("/health", healthHandler.Check)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // Sesuaikan port kamu
	))

	// Grouping untuk route yang butuh Auth
	r.Group(func(r chi.Router) {
		r.Use(auth) // Semua di dalam group ini otomatis kena Auth

		r.Get("/categories", catHandler.GetAll)
		r.Get("/items", itemHandler.GetAllItem)
		r.Get("/items/{id}", itemHandler.GetByID)
		r.Post("/items/create", itemHandler.Create)
		r.Get("/items/export", itemHandler.ExportCSV)
		r.Get("/items/stock-logs", itemHandler.GetStockLogs)

		// Sub-group untuk Admin saja
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin"))
			r.Put("/items/{id}", itemHandler.Update)
			r.Delete("/items/{id}", itemHandler.Delete)
			r.Post("/items/update-stock", itemHandler.UpdateStock)
		})
	})
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 2. Konfigurasi HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.ApiPort,
		Handler:      r, // Masukkan router chi ke sini
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 3. Jalankan Server di Goroutine (Background)
	go func() {
		logger.Info("Server started", "port", cfg.ApiPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server crash", "error", err.Error())
			os.Exit(1)
		}
	}()

	// 4. Tunggu Sinyal Keluar (Stop di sini sampai ada Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// 5. Proses Mematikan (Graceful Shutdown)
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server stopped gracefully")
}
