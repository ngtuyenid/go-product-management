package postgres

import (
	"context"
	"errors"
	"sync"

	"github.com/thanhnguyen/product-api/internal/business/entity"
	"github.com/thanhnguyen/product-api/pkg/logger"
	"gorm.io/gorm"
)

// CategoryRepository implements storage.CategoryRepository
type CategoryRepository struct {
	db     *Database
	logger *logger.Logger
	pool   *sync.Pool
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *Database, logger *logger.Logger) *CategoryRepository {
	return &CategoryRepository{
		db:     db,
		logger: logger,
		pool: &sync.Pool{
			New: func() interface{} {
				return &Category{}
			},
		},
	}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	// Get a model instance from the pool
	model := r.pool.Get().(*Category)
	defer r.pool.Put(model)

	// Reset fields to avoid data leakage
	*model = Category{
		Name:        category.Name,
		Description: category.Description,
	}

	// Create the category
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	// Update the entity with the generated ID
	category.ID = model.ID

	return nil
}

// List lists all categories
func (r *CategoryRepository) List(ctx context.Context) ([]entity.Category, error) {
	var models []Category
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	// Map to entities
	categories := make([]entity.Category, len(models))
	for i, model := range models {
		categories[i] = entity.Category{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
		}
	}

	return categories, nil
}

// FindByID finds a category by ID
func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*entity.Category, error) {
	// Get a model instance from the pool
	model := r.pool.Get().(*Category)
	defer r.pool.Put(model)

	// Find the category
	if err := r.db.WithContext(ctx).First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Map model to entity
	return &entity.Category{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
	}, nil
}

// FindByIDs finds categories by IDs
func (r *CategoryRepository) FindByIDs(ctx context.Context, ids []uint) ([]entity.Category, error) {
	if len(ids) == 0 {
		return []entity.Category{}, nil
	}

	var models []Category
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
		return nil, err
	}

	// Map to entities
	categories := make([]entity.Category, len(models))
	for i, model := range models {
		categories[i] = entity.Category{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
		}
	}

	return categories, nil
}
