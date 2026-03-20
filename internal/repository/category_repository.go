package repository

import (
	"database/sql"
	"fmt"
	"time"
	"warehouse-management-api/internal/entity"
)

// CategoryRepositoryInterface mendefinisikan kontrak method yang harus ada
// Ini digunakan oleh Service agar tidak bergantung langsung pada struct
type CategoryRepositoryInterface interface {
	FindAll() ([]entity.Category, error)
	FindByID(id int) (*entity.Category, error)
	Create(category *entity.Category) error
	Update(category *entity.Category) error
	Delete(id int) error
}

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll() ([]entity.Category, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM categories"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var cat entity.Category
		// Scan ke struct entity
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *CategoryRepository) FindByID(id int) (*entity.Category, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM categories WHERE id = ?"
	row := r.db.QueryRow(query, id)

	var cat entity.Category
	if err := row.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) Create(category *entity.Category) error {
	query := "INSERT INTO categories (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)"
	now := time.Now()
	// Eksekusi query insert
	res, err := r.db.Exec(query, category.Name, category.Description, now, now)
	if err != nil {
		return err
	}

	// Ambil ID yang baru saja digenerate
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Update struct dengan data baru
	category.ID = int(id)
	category.CreatedAt = now
	category.UpdatedAt = now

	return nil
}

func (r *CategoryRepository) Update(category *entity.Category) error {
	query := "UPDATE categories SET name = ?, description = ?, updated_at = ? WHERE id = ?"

	now := time.Now()
	res, err := r.db.Exec(query, category.Name, category.Description, now, category.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("category not found or no changes made")
	}

	return nil
}

func (r *CategoryRepository) Delete(id int) error {
	query := "DELETE FROM categories WHERE id = ?"
	res, err := r.db.Exec(query, id)
	if err != nil {
		// Tips: Jika gagal karena Foreign Key (ada barang yg pakai kategori ini), error akan muncul di sini
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
