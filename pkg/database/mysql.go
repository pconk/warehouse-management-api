package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"warehouse-management-api/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB(cfg *config.DbConfig) (*sql.DB, error) {
	// Format: user:password@tcp(host:port)/dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	log.Println("dsn:", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Pengaturan Connection Pool (Sangat penting untuk performa kerja)
	db.SetMaxIdleConns(10)                 // Jumlah koneksi standby
	db.SetMaxOpenConns(100)                // Maksimal koneksi serentak
	db.SetConnMaxLifetime(1 * time.Hour)   // Umur maksimal koneksi
	db.SetConnMaxIdleTime(5 * time.Minute) // Waktu idle sebelum diputus

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
