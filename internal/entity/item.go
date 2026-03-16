package entity

import "time"

type Item struct {
	ID         int       `json:"id"`
	CategoryID int       `json:"category_id" validate:"required"`
	SKU        string    `json:"sku" validate:"required,min=3"`
	Name       string    `json:"name" validate:"required"`
	Price      float64   `json:"price" validate:"required,gt=0"`
	Stock      int       `json:"stock" validate:"min=0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Opsional: Jika ingin join data kategori
	CategoryName string     `json:"category_name,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type UpdateItemRequest struct {
	CategoryID int     `json:"category_id" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	Price      float64 `json:"price" validate:"required,gt=0"`
}

type UpdateStockRequest struct {
	ItemID   int    `json:"item_id" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=IN OUT"`
	Quantity int    `json:"quantity" validate:"required,gt=0"`
	Reason   string `json:"reason"`
}
