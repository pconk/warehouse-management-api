package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DbConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type AppConfig struct {
	Db                  DbConfig
	ApiPort             string
	RedisAddress        string
	WarehouseID         string
	QueueName           string
	EnableLowStockAlert bool
	LowStockThreshold   int
	LowStockAlertEmail  string
	AuditServiceURL     string
	JWTSecret           string
}

func LoadConfig() (*AppConfig, error) {
	godotenv.Load() // Abaikan error jika file .env tidak ada

	cfg := &AppConfig{
		Db: DbConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "warehouse_db"),
		},
		ApiPort:             getEnv("API_PORT", "8080"),
		RedisAddress:        getEnv("REDIS_ADDRESS", "localhost:6379"),
		WarehouseID:         getEnv("WAREHOUSE_ID", "WH_001"),
		QueueName:           getEnv("QUEUE_NAME", "email_jobs"),
		AuditServiceURL:     getEnv("AUDIT_SERVICE_URL", "localhost:50051"),
		JWTSecret:           getEnv("JWT_SECRET", "rahasia-super-aman"),
		EnableLowStockAlert: getEnv("ENABLE_LOW_STOCK_ALERT", "true") == "true",
		LowStockThreshold:   getEnvInt("LOW_STOCK_THRESHOLD", 5),
		LowStockAlertEmail:  getEnv("LOW_STOCK_ALERT_EMAIL", "admin@example.com"),
	}
	return cfg, nil
}

// getEnv membaca environment variable atau mengembalikan nilai fallback.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return fallback
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}
