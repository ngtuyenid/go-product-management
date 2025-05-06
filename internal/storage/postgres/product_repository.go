package postgres

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/thanhnguyen/product-api/internal/business/entity"
	"github.com/thanhnguyen/product-api/pkg/logger"
	"gorm.io/gorm"
)

// ProductRepository implements storage.ProductRepository
type ProductRepository struct {
	db           *Database
	logger       *logger.Logger
	productPool  *sync.Pool
	categoryPool *sync.Pool
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *Database, logger *logger.Logger) *ProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger,
		productPool: &sync.Pool{
			New: func() interface{} {
				return &Product{}
			},
		},
		categoryPool: &sync.Pool{
			New: func() interface{} {
				return &Category{}
			},
		},
	}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) error {
	// Get a model instance from the pool
	model := r.productPool.Get().(*Product)
	defer r.productPool.Put(model)

	// Reset fields to avoid data leakage
	*model = Product{
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
		Status:        product.Status,
	}

	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the product
	if err := tx.Create(model).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Add categories
	if len(product.Categories) > 0 {
		for _, cat := range product.Categories {
			if err := tx.Exec("INSERT INTO product_categories (product_id, category_id) VALUES (?, ?)", model.ID, cat.ID).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Update the entity with the generated ID
	product.ID = model.ID
	product.CreatedAt = model.CreatedAt
	product.UpdatedAt = model.UpdatedAt

	return nil
}

// List lists products with filtering and pagination
func (r *ProductRepository) List(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, int64, error) {
	var (
		products []Product
		count    int64
		wg       sync.WaitGroup
		countErr error
		listErr  error
		mu       sync.Mutex
	)

	// Build query
	query := r.db.WithContext(ctx).Model(&Product{})

	// Apply filters
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if filter.CategoryID != 0 {
		query = query.Joins("JOIN product_categories pc ON products.id = pc.product_id").
			Where("pc.category_id = ?", filter.CategoryID)
	}

	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	// Count total in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := query
		if countErr = q.Count(&count).Error; countErr != nil {
			r.logger.WithError(countErr).Error("Failed to count products")
		}
	}()

	// Apply pagination
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// Apply sorting
	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query = query.Order(filter.SortBy + " " + order)
	} else {
		query = query.Order("id DESC")
	}

	// Get products in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := query
		if listErr = q.Offset(offset).Limit(pageSize).Find(&products).Error; listErr != nil {
			r.logger.WithError(listErr).Error("Failed to list products")
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	// Check for errors
	if countErr != nil {
		return nil, 0, countErr
	}
	if listErr != nil {
		return nil, 0, listErr
	}

	// Map to entities with goroutines
	result := make([]entity.Product, len(products))
	if len(products) > 0 {
		wg = sync.WaitGroup{}
		wg.Add(len(products))

		for i, p := range products {
			go func(i int, p Product) {
				defer wg.Done()

				// Map product
				product := entity.Product{
					ID:            p.ID,
					Name:          p.Name,
					Description:   p.Description,
					Price:         p.Price,
					StockQuantity: p.StockQuantity,
					Status:        p.Status,
					CreatedAt:     p.CreatedAt,
					UpdatedAt:     p.UpdatedAt,
				}

				// Get categories
				var categories []Category
				if err := r.db.WithContext(ctx).Model(&p).Association("Categories").Find(&categories); err == nil {
					for _, c := range categories {
						product.Categories = append(product.Categories, entity.Category{
							ID:          c.ID,
							Name:        c.Name,
							Description: c.Description,
						})
					}
				}

				// Store in result
				mu.Lock()
				result[i] = product
				mu.Unlock()
			}(i, p)
		}

		wg.Wait()
	}

	return result, count, nil
}

// FindByID finds a product by ID
func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	// Get a model instance from the pool
	model := r.productPool.Get().(*Product)
	defer r.productPool.Put(model)

	// Find the product
	if err := r.db.WithContext(ctx).First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Map model to entity
	product := &entity.Product{
		ID:            model.ID,
		Name:          model.Name,
		Description:   model.Description,
		Price:         model.Price,
		StockQuantity: model.StockQuantity,
		Status:        model.Status,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}

	// Get categories
	var categories []Category
	if err := r.db.WithContext(ctx).Model(model).Association("Categories").Find(&categories); err == nil {
		for _, c := range categories {
			product.Categories = append(product.Categories, entity.Category{
				ID:          c.ID,
				Name:        c.Name,
				Description: c.Description,
			})
		}
	}

	return product, nil
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *entity.Product) error {
	// Get a model instance from the pool
	model := r.productPool.Get().(*Product)
	defer r.productPool.Put(model)

	// Find the product
	if err := r.db.WithContext(ctx).First(model, product.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// Update fields
	model.Name = product.Name
	model.Description = product.Description
	model.Price = product.Price
	model.StockQuantity = product.StockQuantity
	model.Status = product.Status

	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the product
	if err := tx.Save(model).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update categories if provided
	if len(product.Categories) > 0 {
		// Remove existing categories
		if err := tx.Exec("DELETE FROM product_categories WHERE product_id = ?", model.ID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Add new categories
		for _, cat := range product.Categories {
			if err := tx.Exec("INSERT INTO product_categories (product_id, category_id) VALUES (?, ?)", model.ID, cat.ID).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Update the entity
	product.UpdatedAt = model.UpdatedAt

	return nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Product{}, id).Error
}

// AddCategories adds categories to a product
func (r *ProductRepository) AddCategories(ctx context.Context, productID uint, categoryIDs []uint) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, categoryID := range categoryIDs {
		if err := tx.Exec("INSERT INTO product_categories (product_id, category_id) VALUES (?, ?)", productID, categoryID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
