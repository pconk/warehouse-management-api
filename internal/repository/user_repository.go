package repository

import (
	"database/sql"
	"fmt"
	"warehouse-management-api/internal/entity"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByApiKey(apiKey string) (*entity.User, error) {
	var user entity.User
	query := "SELECT id, username, api_key, role FROM users WHERE api_key = ?"

	// Langsung .Scan(), tidak perlu .Next()
	err := r.db.QueryRow(query, apiKey).Scan(&user.ID, &user.Username, &user.ApiKey, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}
