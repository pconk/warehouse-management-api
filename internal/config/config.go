// config/config.go
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type AppConfig struct {
	ApiKey              string
	ApiPort             string
	Db                  *DBConfig
	RedisAddress        string
	EnableLowStockAlert bool
	LowStockAlertEmail  string
	QueueName           string
}

func LoadConfig() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	lowStockAlertEmail := os.Getenv("LOW_STOCK_ALERT_EMAIL")
	if lowStockAlertEmail == "" {
		lowStockAlertEmail = "admin@warehouse.com"
	}
	return &AppConfig{
		ApiKey:  os.Getenv("API_KEY"),
		ApiPort: os.Getenv("APP_PORT"),
		Db: &DBConfig{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
		},
		EnableLowStockAlert: os.Getenv("ENABLE_LOW_STOCK_ALERT") == "true",
		LowStockAlertEmail:  lowStockAlertEmail,
		RedisAddress:        os.Getenv("REDIS_ADDRESS"),
		QueueName:           os.Getenv("QUEUE_NAME"),
	}, nil
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
