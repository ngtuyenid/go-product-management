package dto

import (
	"time"

	"github.com/thanhnguyen/product-api/internal/business/entity"
)

// ProductRequest represents a request to create or update a product
type ProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	StockQuantity int     `json:"stock_quantity" binding:"required,gte=0"`
	CategoryIDs   []uint  `json:"category_ids" binding:"required"`
}

// ProductResponse represents a product in the response
type ProductResponse struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	StockQuantity int      `json:"stock_quantity"`
	Status        string   `json:"status"`
	Categories    []string `json:"categories"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ProductListRequest represents a request to list products
type ProductListRequest struct {
	Search     string   `form:"search"`
	Page       int      `form:"page,default=1"`
	PageSize   int      `form:"page_size,default=10"`
	CategoryID uint     `form:"category_id"`
	MinPrice   *float64 `form:"min_price"`
	MaxPrice   *float64 `form:"max_price"`
	SortBy     string   `form:"sort_by"`
	SortOrder  string   `form:"sort_order"`
}

// ProductListResponse represents a paginated list of products
type ProductListResponse struct {
	Items      []ProductResponse `json:"items"`
	TotalItems int64             `json:"total_items"`
	TotalPages int               `json:"total_pages"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// ToEntity converts a ProductRequest to an entity.Product
func (r *ProductRequest) ToEntity() *entity.Product {
	return &entity.Product{
		Name:          r.Name,
		Description:   r.Description,
		Price:         r.Price,
		StockQuantity: r.StockQuantity,
		Status:        "active", // Default status
	}
}

// ToProductFilter converts a ProductListRequest to an entity.ProductFilter
func (r *ProductListRequest) ToProductFilter() entity.ProductFilter {
	return entity.ProductFilter{
		Search:     r.Search,
		Page:       r.Page,
		PageSize:   r.PageSize,
		CategoryID: r.CategoryID,
		MinPrice:   r.MinPrice,
		MaxPrice:   r.MaxPrice,
		SortBy:     r.SortBy,
		SortOrder:  r.SortOrder,
	}
}

// FromEntity converts an entity.Product to a ProductResponse
func FromEntity(p entity.Product) ProductResponse {
	// Extract category names
	categories := make([]string, 0, len(p.Categories))
	for _, c := range p.Categories {
		categories = append(categories, c.Name)
	}

	return ProductResponse{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		StockQuantity: p.StockQuantity,
		Status:        p.Status,
		Categories:    categories,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}
