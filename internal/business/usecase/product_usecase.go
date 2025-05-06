package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/thanhnguyen/product-api/internal/business/entity"
	"github.com/thanhnguyen/product-api/internal/storage"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// ProductUseCase defines the product business logic
type ProductUseCase interface {
	CreateProduct(ctx context.Context, product *entity.Product, categoryIDs []uint) error
	ListProducts(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, int64, error)
	GetProduct(ctx context.Context, id uint) (*entity.Product, error)
	UpdateProduct(ctx context.Context, product *entity.Product, categoryIDs []uint) error
	DeleteProduct(ctx context.Context, id uint) error
}

// productUseCase implements ProductUseCase
type productUseCase struct {
	productRepo  storage.ProductRepository
	categoryRepo storage.CategoryRepository
	logger       *logger.Logger
	cacheTimeout time.Duration
}

// NewProductUseCase creates a new ProductUseCase
func NewProductUseCase(
	productRepo storage.ProductRepository,
	categoryRepo storage.CategoryRepository,
	logger *logger.Logger,
	cacheTimeout time.Duration,
) ProductUseCase {
	return &productUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		logger:       logger,
		cacheTimeout: cacheTimeout,
	}
}

// CreateProduct creates a new product
func (uc *productUseCase) CreateProduct(ctx context.Context, product *entity.Product, categoryIDs []uint) error {
	// Validate product
	if err := validateProduct(product); err != nil {
		return err
	}

	// Get categories
	if len(categoryIDs) > 0 {
		categories, err := uc.categoryRepo.FindByIDs(ctx, categoryIDs)
		if err != nil {
			return err
		}
		if len(categories) != len(categoryIDs) {
			return errors.New("one or more categories not found")
		}
		product.Categories = categories
	}

	// Set default status if not provided
	if product.Status == "" {
		product.Status = "active"
	}

	// Create product
	return uc.productRepo.Create(ctx, product)
}

// ListProducts lists products with filtering and pagination
func (uc *productUseCase) ListProducts(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, int64, error) {
	// Set default values for pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	// Get products from repository
	return uc.productRepo.List(ctx, filter)
}

// GetProduct gets a product by ID
func (uc *productUseCase) GetProduct(ctx context.Context, id uint) (*entity.Product, error) {
	product, err := uc.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	return product, nil
}

// UpdateProduct updates a product
func (uc *productUseCase) UpdateProduct(ctx context.Context, product *entity.Product, categoryIDs []uint) error {
	// Check if product exists
	existingProduct, err := uc.productRepo.FindByID(ctx, product.ID)
	if err != nil {
		return err
	}
	if existingProduct == nil {
		return errors.New("product not found")
	}

	// Validate product
	if err := validateProduct(product); err != nil {
		return err
	}

	// Get categories if provided
	if len(categoryIDs) > 0 {
		categories, err := uc.categoryRepo.FindByIDs(ctx, categoryIDs)
		if err != nil {
			return err
		}
		if len(categories) != len(categoryIDs) {
			return errors.New("one or more categories not found")
		}
		product.Categories = categories
	}

	// Update product
	return uc.productRepo.Update(ctx, product)
}

// DeleteProduct deletes a product
func (uc *productUseCase) DeleteProduct(ctx context.Context, id uint) error {
	// Check if product exists
	product, err := uc.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}

	// Delete product
	return uc.productRepo.Delete(ctx, id)
}

// validateProduct validates a product
func validateProduct(product *entity.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Price <= 0 {
		return errors.New("product price must be greater than zero")
	}
	if product.StockQuantity < 0 {
		return errors.New("product stock quantity cannot be negative")
	}
	return nil
}
