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
	"warehouse-management-api/internal/queue"
	"warehouse-management-api/internal/repository"
	"warehouse-management-api/internal/service"
	"warehouse-management-api/pkg/database"
	"warehouse-management-api/pkg/redis"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ForwardRequestIDInterceptor otomatis meneruskan request_id dari Chi ke gRPC Metadata
func ForwardRequestIDInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqID := middleware.GetRequestID(ctx)
		if reqID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", reqID)
		}

		// Ambil token dari context (disuntikkan oleh AuthMiddleware)
		if token := middleware.GetToken(ctx); token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// @title Warehouse Management API
// @version 1.0
// @description API untuk manajemen stok gudang.
// @description
// @description ## Cara Mendapatkan Token JWT untuk Testing:
// @description 1. Pastikan **Auth Service** berjalan (Port 8081).
// @description 2. Lakukan login melalui endpoint POST `http://localhost:8081/auth/login`.
// @description 3. Copy string token dari response JSON.
// @description 4. Klik tombol **Authorize** di kanan atas, lalu masukkan format: `Bearer [token_kamu]`.
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
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token JWT dengan format: Bearer [your-token]
func main() {
	// Setup JSON Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("configuration file not found")
		os.Exit(1)
	}
	logger.Info("print env", "cfg", cfg)

	if cfg.RedisAddress == "" {
		logger.Error("REDIS_ADDRESS is invalid")
		os.Exit(1)
	}

	// Koneksi DB
	db, err := database.ConnectDB(&cfg.Db)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Init redis
	rdb := redis.NewRedisClient(cfg.RedisAddress)
	emailProducer := queue.NewEmailProducer(rdb, cfg.QueueName)

	// Init Audit Service Client (gRPC)
	conn, err := grpc.NewClient(
		cfg.AuditServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(ForwardRequestIDInterceptor()),
	)
	if err != nil {
		logger.Error("Failed to connect to Audit Service", "error", err)
		// Pertimbangkan untuk os.Exit(1) di sini jika koneksi ke audit service wajib
	}
	auditClient := service.NewAuditClient(conn, logger)

	// Init Repository
	healthRepo := repository.NewHealthRepository(db)
	itemRepo := repository.NewItemRepository(db)
	catRepo := repository.NewCategoryRepository(db)

	// Init service
	itemService := service.NewItemService(itemRepo, emailProducer, auditClient, logger, cfg)
	categoryService := service.NewCategoryService(catRepo, logger)

	// Init Handler (Sambil lempar repo-nya)
	healthHandler := handler.NewHealthHandler(healthRepo, logger)
	itemHandler := handler.NewItemHandler(itemService, logger)
	catHandler := handler.NewCategoryHandler(categoryService, logger)

	// Inisialisasi Middleware dengan API Key dari .env
	auth := middleware.AuthMiddleware(logger, cfg.JWTSecret)
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 1. Setup Router Chi
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)

	// Middleware global (untuk semua route)
	r.Use(middleware.Logger(logger))

	r.Get("/health", healthHandler.Check)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+cfg.ApiPort+"/swagger/doc.json"), // Sesuaikan port kamu
	))

	// Grouping untuk route yang butuh Auth
	r.Group(func(r chi.Router) {
		r.Use(auth) // Semua di dalam group ini otomatis kena Auth

		r.Get("/categories", catHandler.GetAll)
		r.Get("/categories/{id}", catHandler.GetByID)
		r.Get("/items", itemHandler.GetAllItem)
		r.Get("/items/{id}", itemHandler.GetByID)
		r.Post("/items/create", itemHandler.Create)
		r.Get("/items/export", itemHandler.ExportCSV)
		r.Get("/items/stock-logs", itemHandler.GetStockLogs)

		// Sub-group untuk Admin saja
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin"))
			r.Put("/categories/{id}", catHandler.Update)
			r.Delete("/categories/{id}", catHandler.Delete)
			r.Post("/categories", catHandler.Create)

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
