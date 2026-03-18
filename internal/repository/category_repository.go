package repository

import (
	"database/sql"
	"warehouse-management-api/internal/entity"
)

type CategoryRepository struct {
	db *sql.DB
}

type CategoryRepositoryInterface interface {
	FindAll(limit, offset int) ([]entity.Category, int64, error)
}

func NewCategoryRepository(db *sql.DB) CategoryRepositoryInterface {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll(limit, offset int) ([]entity.Category, int64, error) {
	// 1. Query Data
	query := "SELECT id, name, description, created_at FROM categories ORDER BY name ASC LIMIT ? OFFSET ?"
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	categories := []entity.Category{}
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		categories = append(categories, c)
	}

	// 2. Query Total
	var total int64
	r.db.QueryRow("SELECT COUNT(id) FROM categories").Scan(&total)

	return categories, total, nil
}
