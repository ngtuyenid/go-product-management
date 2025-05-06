package entity

// CategoryStat represents statistics for a category
type CategoryStat struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	ProductCount int    `json:"product_count"`
}

// WishlistStat represents statistics for a product in wishlists
type WishlistStat struct {
	ProductID     uint   `json:"product_id"`
	ProductName   string `json:"product_name"`
	WishlistCount int    `json:"wishlist_count"`
}

// TopProduct represents a top product by some metric
type TopProduct struct {
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	Count       int    `json:"count"`
	Metric      string `json:"metric"`
}
