package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/thanhnguyen/product-api/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Database represents a connection to the database
type Database struct {
	*gorm.DB
	logger *logger.Logger
}

// Config contains database configuration
type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

// NewPostgresDB creates a new database connection
func NewPostgresDB(dsn string, maxOpenConns, minOpenConns int, timeout time.Duration) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set connection pool limits
	sqlDB.SetMaxIdleConns(minOpenConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(timeout)

	return &Database{
		DB: db,
	}, nil
}

// WithContext returns a GORM DB instance with the given context
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate migrates the database schema
func (d *Database) AutoMigrate() error {
	d.logger.Info("Auto-migrating database schema")
	err := d.DB.AutoMigrate(
		&User{},
		&Product{},
		&Category{},
		&Review{},
		&Wishlist{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}
	d.logger.Info("Database migration completed")
	return nil
}

// Seed seeds the database with initial data
func (db *Database) Seed() error {
	db.logger.Info("Seeding database with initial data")

	// Check if admin user exists
	var adminCount int64
	db.DB.Model(&User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount == 0 {
		admin := User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: "$2a$10$aeFCjbHcgJjK.ZBbrNk.pO4H4SCNPVpqG8ZlGI.aO7xFb9l/o9bqm", // admin123
			FullName:     "Admin User",
			Role:         "admin",
		}
		if err := db.DB.Create(&admin).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		db.logger.Info("Admin user created")
	}

	// Check if categories exist
	var categoryCount int64
	db.DB.Model(&Category{}).Count(&categoryCount)
	if categoryCount == 0 {
		categories := []Category{
			{Name: "Electronics", Description: "Electronic devices and gadgets"},
			{Name: "Clothing", Description: "Clothing and apparel"},
			{Name: "Books", Description: "Books and publications"},
			{Name: "Home", Description: "Home and garden products"},
			{Name: "Sports", Description: "Sports and outdoor equipment"},
		}
		if err := db.DB.Create(&categories).Error; err != nil {
			return fmt.Errorf("failed to create categories: %w", err)
		}
		db.logger.Info("Categories created")
	}

	db.logger.Info("Database seeding completed")
	return nil
}
