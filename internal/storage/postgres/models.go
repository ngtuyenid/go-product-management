package postgres

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the database
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;size:255;not null"`
	Email        string    `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	FullName     string    `gorm:"size:255"`
	Role         string    `gorm:"size:50;default:user"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// Product represents a product in the database
type Product struct {
	ID            uint    `gorm:"primaryKey"`
	Name          string  `gorm:"size:255;not null"`
	Description   string  `gorm:"type:text"`
	Price         float64 `gorm:"type:decimal(10,2)"`
	StockQuantity int
	Status        string     `gorm:"size:50;default:active"`
	Categories    []Category `gorm:"many2many:product_categories;"`
	Reviews       []Review   `gorm:"foreignKey:ProductID"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
}

// Category represents a product category in the database
type Category struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	Products    []Product `gorm:"many2many:product_categories;"`
}

// Review represents a product review in the database
type Review struct {
	ID        uint      `gorm:"primaryKey"`
	ProductID uint      `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   string    `gorm:"type:text"`
	User      User      `gorm:"foreignKey:UserID"`
	Product   Product   `gorm:"foreignKey:ProductID"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// Wishlist represents a product in a user's wishlist in the database
type Wishlist struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false"`
	ProductID uint      `gorm:"primaryKey;autoIncrement:false"`
	AddedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	User      User      `gorm:"foreignKey:UserID"`
	Product   Product   `gorm:"foreignKey:ProductID"`
}

// TableNames
func (User) TableName() string {
	return "users"
}

func (Product) TableName() string {
	return "products"
}

func (Category) TableName() string {
	return "categories"
}

func (Review) TableName() string {
	return "reviews"
}

func (Wishlist) TableName() string {
	return "wishlist"
}

// BeforeCreate hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	return nil
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = "active"
	}
	return nil
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.Rating < 1 {
		r.Rating = 1
	} else if r.Rating > 5 {
		r.Rating = 5
	}
	return nil
}
