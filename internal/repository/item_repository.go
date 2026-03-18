package repository

import (
	"database/sql"
	"fmt"
	"warehouse-management-api/internal/entity"
)

type ItemRepository struct {
	db *sql.DB
}

type ItemRepositoryInterface interface {
	FindByID(id int) (*entity.Item, error)
	List(limit, offset int, name string, categoryID int) ([]entity.Item, int64, error)
	FindAllForExport() ([]entity.Item, error)
	Insert(item entity.Item) error
	Update(id int, item entity.UpdateItemRequest) error
	UpdateStock(req entity.UpdateStockRequest, userID int) error
	Delete(id int) error
	GetStockLogs(limit, offset int, itemID int, logType string) ([]entity.StockLog, int64, error)
}

func NewItemRepository(db *sql.DB) ItemRepositoryInterface {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) CountAll(name string, categoryID int) (int64, error) {
	query := "SELECT COUNT(id) FROM items WHERE deleted_at IS NULL"
	var args []interface{}

	if name != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+name+"%")
	}
	if categoryID > 0 {
		query += " AND category_id = ?"
		args = append(args, categoryID)
	}

	var count int64
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (r *ItemRepository) FindByID(id int) (*entity.Item, error) {
	var item entity.Item
	query := `
		SELECT i.id, i.category_id, i.sku, i.name, i.price, i.stock, i.created_at, i.updated_at, c.name as category_name
		FROM items i
		INNER JOIN categories c ON i.category_id = c.id
		WHERE i.deleted_at IS NULL AND i.id = ?`

	err := r.db.QueryRow(query, id).Scan(
		&item.ID, &item.CategoryID, &item.SKU, &item.Name, &item.Price,
		&item.Stock, &item.CreatedAt, &item.UpdatedAt, &item.CategoryName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Mengembalikan nil agar handler tahu datanya memang tidak ada
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (r *ItemRepository) FindAll(limit, offset int, name string, categoryID int) ([]entity.Item, error) {
	query := `
	SELECT i.id, i.category_id, i.sku, i.name, i.price, i.stock, i.created_at, i.updated_at, c.name as category_name
	FROM items i
	INNER JOIN categories c ON i.category_id = c.id 
	WHERE deleted_at IS NULL`

	var args []interface{}

	// Filter Nama (Partial Search)
	if name != "" {
		query += " AND i.name LIKE ?"
		args = append(args, "%"+name+"%")
	}

	// Filter Kategori
	if categoryID > 0 {
		query += " AND i.category_id = ?"
		args = append(args, categoryID)
	}

	// Tambahkan Order, Limit, dan Offset
	query += " ORDER BY i.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []entity.Item{}
	for rows.Next() {
		var item entity.Item

		err := rows.Scan(&item.ID, &item.CategoryID, &item.SKU, &item.Name, &item.Price, &item.Stock, &item.CreatedAt, &item.UpdatedAt, &item.CategoryName)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepository) FindAllForExport() ([]entity.Item, error) {
	query := `
	SELECT i.id, i.category_id, i.sku, i.name, i.price, i.stock, i.created_at, i.updated_at, c.name as category_name
	FROM items i
	INNER JOIN categories c ON i.category_id = c.id
	WHERE i.deleted_at IS NULL`

	rows, err := r.db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.Item
	for rows.Next() {
		var item entity.Item
		err := rows.Scan(&item.ID, &item.CategoryID, &item.SKU, &item.Name, &item.Price, &item.Stock, &item.CreatedAt, &item.UpdatedAt, &item.CategoryName)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// Ambil semua daftar barang beserta stoknya
func (r *ItemRepository) List(limit, offset int, name string, categoryID int) ([]entity.Item, int64, error) {
	// Query 1: Ambil Data
	items, err := r.FindAll(limit, offset, name, categoryID)
	if err != nil {
		return nil, 0, err
	}

	// Query 2: Ambil Total
	total, err := r.CountAll(name, categoryID)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ItemRepository) Insert(item entity.Item) error {
	query := "INSERT INTO items (category_id, sku, name, price, stock) VALUES  (?, ?, ?, ?, ?)"

	_, err := r.db.Exec(query, item.CategoryID, item.SKU, item.Name, item.Price, item.Stock)
	if err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) Update(id int, item entity.UpdateItemRequest) error {
	query := `
		UPDATE items 
		SET category_id = ?, name = ?, price = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`

	res, err := r.db.Exec(query, item.CategoryID, item.Name, item.Price, id)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found or no changes made")
	}
	return nil
}

func (r *ItemRepository) UpdateStock(req entity.UpdateStockRequest, userID int) error {
	// 1. Mulai Transaksi
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Helper untuk rollback jika terjadi error di tengah jalan
	defer tx.Rollback()

	// 2. Update stok di tabel items
	var queryUpdate string
	if req.Type == "IN" {
		queryUpdate = "UPDATE items SET stock = stock + ? WHERE id = ? AND deleted_at IS NULL"
	} else {
		queryUpdate = "UPDATE items SET stock = stock - ? WHERE id = ? AND deleted_at IS NULL AND stock >= ?"
	}

	var res sql.Result
	if req.Type == "OUT" {
		res, err = tx.Exec(queryUpdate, req.Quantity, req.ItemID, req.Quantity)
	} else {
		res, err = tx.Exec(queryUpdate, req.Quantity, req.ItemID)
	}

	if err != nil {
		return err
	}

	// Cek apakah item_id ditemukan
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found or insufficient stock")
	}

	// 3. Catat riwayat ke tabel stock_logs
	queryLog := "INSERT INTO stock_logs (item_id, user_id, type, quantity, reason) VALUES (?, ?, ?, ?, ?)"
	_, err = tx.Exec(queryLog, req.ItemID, userID, req.Type, req.Quantity, req.Reason)
	if err != nil {
		return err
	}

	// 4. Commit jika semua lancar
	return tx.Commit()
}

func (r *ItemRepository) Delete(id int) error {
	query := "UPDATE items SET deleted_at = NOW() WHERE id = ?"
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}

func (r *ItemRepository) GetStockLogs(limit, offset int, itemID int, logType string) ([]entity.StockLog, int64, error) {
	// 1. Query Data dengan Join
	query := `
		SELECT sl.id, sl.item_id, i.name, i.sku, sl.type, sl.quantity, sl.reason, sl.created_at, u.username
	FROM stock_logs sl
		INNER JOIN items i ON sl.item_id = i.id
		LEFT JOIN users u ON sl.user_id = u.id
		WHERE 1=1`
	var args []interface{}
	if itemID > 0 {
		query += " AND sl.item_id = ?"
		args = append(args, itemID)
	}
	if logType != "" {
		query += " AND sl.type = ?"
		args = append(args, logType)
	}

	query += " ORDER BY sl.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := []entity.StockLog{}
	for rows.Next() {
		var l entity.StockLog

		if err := rows.Scan(&l.ID, &l.ItemID, &l.ItemName, &l.ItemSKU, &l.Type, &l.Quantity, &l.Reason, &l.CreatedAt, &l.UserName); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	// 2. Query Total untuk Pagination Meta
	var total int64
	countQuery := "SELECT COUNT(*) FROM stock_logs WHERE 1=1"

	if itemID > 0 {
		countQuery += " AND item_id = ?"
	}
	if logType != "" {
		countQuery += " AND type = ?"
	}

	if err := r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
