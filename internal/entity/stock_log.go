package entity

import "time"

type StockLog struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	ItemName  string    `json:"item_name,omitempty"` // Dari Join
	ItemSKU   string    `json:"item_sku,omitempty"`  // Dari Join
	Type      string    `json:"type"`                // IN / OUT
	Quantity  int       `json:"quantity"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int64     `json:"user_id,omitempty"`
	UserName  *string   `json:"user_name,omitempty"`
}
