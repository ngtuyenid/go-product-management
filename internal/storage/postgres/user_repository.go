package postgres

import (
	"context"
	"errors"
	"sync"

	"github.com/thanhnguyen/product-api/internal/business/entity"
	"github.com/thanhnguyen/product-api/pkg/logger"
	"gorm.io/gorm"
)

// UserRepository implements storage.UserRepository
type UserRepository struct {
	db     *Database
	logger *logger.Logger
	pool   *sync.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *Database, logger *logger.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
		pool: &sync.Pool{
			New: func() interface{} {
				return &User{}
			},
		},
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	// Get a model instance from the pool
	model := r.pool.Get().(*User)
	defer r.pool.Put(model)

	// Reset fields to avoid data leakage
	*model = User{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		Role:         user.Role,
	}

	// Create the user
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	// Update the entity with the generated ID
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	// Get a model instance from the pool
	model := r.pool.Get().(*User)
	defer r.pool.Put(model)

	// Find the user
	if err := r.db.WithContext(ctx).First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Map model to entity
	return &entity.User{
		ID:           model.ID,
		Username:     model.Username,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		FullName:     model.FullName,
		Role:         model.Role,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	// Get a model instance from the pool
	model := r.pool.Get().(*User)
	defer r.pool.Put(model)

	// Find the user
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Map model to entity
	return &entity.User{
		ID:           model.ID,
		Username:     model.Username,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		FullName:     model.FullName,
		Role:         model.Role,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	// Get a model instance from the pool
	model := r.pool.Get().(*User)
	defer r.pool.Put(model)

	// Find the user
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Map model to entity
	return &entity.User{
		ID:           model.ID,
		Username:     model.Username,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		FullName:     model.FullName,
		Role:         model.Role,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	// Get a model instance from the pool
	model := r.pool.Get().(*User)
	defer r.pool.Put(model)

	// Find the user
	if err := r.db.WithContext(ctx).First(model, user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// Update fields
	model.Username = user.Username
	model.Email = user.Email
	model.PasswordHash = user.PasswordHash
	model.FullName = user.FullName
	model.Role = user.Role

	// Save the user
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}

	// Update the entity
	user.UpdatedAt = model.UpdatedAt

	return nil
}
