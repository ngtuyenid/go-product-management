package storage

import (
	"context"

	"github.com/thanhnguyen/product-api/internal/business/entity"
)

// UserRepository defines methods for user storage operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

// ProductRepository defines methods for product storage operations
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	List(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uint) error
	AddCategories(ctx context.Context, productID uint, categoryIDs []uint) error
}

// CategoryRepository defines methods for category storage operations
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	List(ctx context.Context) ([]entity.Category, error)
	FindByID(ctx context.Context, id uint) (*entity.Category, error)
	FindByIDs(ctx context.Context, ids []uint) ([]entity.Category, error)
}

// ReviewRepository defines methods for review storage operations
type ReviewRepository interface {
	Create(ctx context.Context, review *entity.Review) error
	List(ctx context.Context, productID uint) ([]entity.Review, error)
	FindByID(ctx context.Context, id uint) (*entity.Review, error)
}

// WishlistRepository defines methods for wishlist storage operations
type WishlistRepository interface {
	Add(ctx context.Context, userID, productID uint) error
	Remove(ctx context.Context, userID, productID uint) error
	List(ctx context.Context, userID uint) ([]entity.Product, error)
	IsProductInWishlist(ctx context.Context, userID, productID uint) (bool, error)
}
