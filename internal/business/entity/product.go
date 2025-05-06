package entity

import "time"

// Product represents a product in the system
type Product struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Price         float64    `json:"price"`
	StockQuantity int        `json:"stock_quantity"`
	Status        string     `json:"status"`
	Categories    []Category `json:"categories,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ProductFilter contains filtering criteria for products
type ProductFilter struct {
	Search     string   `json:"search"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	CategoryID uint     `json:"category_id,omitempty"`
	MinPrice   *float64 `json:"min_price,omitempty"`
	MaxPrice   *float64 `json:"max_price,omitempty"`
	SortBy     string   `json:"sort_by,omitempty"`
	SortOrder  string   `json:"sort_order,omitempty"`
}
