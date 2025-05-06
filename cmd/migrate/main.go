package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Migration represents a single migration
type Migration struct {
	Name string
	Path string
	Type string // "up" or "down"
}

func main() {
	// Parse command line arguments
	var down bool
	var migrationID string
	var envFile string

	flag.BoolVar(&down, "down", false, "Roll back migrations instead of applying them")
	flag.StringVar(&migrationID, "migration", "", "Specify a specific migration to run (optional)")
	flag.StringVar(&envFile, "env", ".env", "Path to the .env file")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to the database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	// Create migrations table if it doesn't exist
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get applied migrations
	var appliedMigrations []string
	err = db.Table("migrations").Pluck("name", &appliedMigrations).Error
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	// Get available migrations
	migrations, err := loadMigrations("migrations/sql", down)
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// Filter migrations
	var migrationsToApply []Migration
	if down {
		// Sort in reverse order for down migrations
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Name > migrations[j].Name
		})

		// Only consider applied migrations for rollback
		for _, migration := range migrations {
			if contains(appliedMigrations, migration.Name) {
				if migrationID == "" || migration.Name == migrationID {
					migrationsToApply = append(migrationsToApply, migration)
				}
			}
		}
	} else {
		// Sort in ascending order for up migrations
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Name < migrations[j].Name
		})

		// Only apply migrations that haven't been applied yet
		for _, migration := range migrations {
			if !contains(appliedMigrations, migration.Name) {
				if migrationID == "" || migration.Name == migrationID {
					migrationsToApply = append(migrationsToApply, migration)
				}
			}
		}
	}

	if len(migrationsToApply) == 0 {
		log.Println("No migrations to apply")
		return
	}

	// Apply migrations
	for _, migration := range migrationsToApply {
		log.Printf("Applying migration: %s\n", migration.Name)

		// Read migration file
		content, err := ioutil.ReadFile(migration.Path)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", migration.Path, err)
		}

		// Begin transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.Fatalf("Failed to begin transaction: %v", tx.Error)
		}

		// Execute migration
		err = tx.Exec(string(content)).Error
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration %s: %v", migration.Name, err)
		}

		// Update migrations table
		if down {
			err = tx.Exec("DELETE FROM migrations WHERE name = $1", migration.Name).Error
		} else {
			err = tx.Exec("INSERT INTO migrations (name) VALUES ($1)", migration.Name).Error
		}

		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to update migrations table: %v", err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}

		log.Printf("Successfully applied migration: %s\n", migration.Name)
	}

	log.Println("Migrations completed successfully")
}

// loadMigrations loads all migration files from the specified directory
func loadMigrations(dir string, down bool) ([]Migration, error) {
	var migrations []Migration

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := filepath.Base(path)
		if !strings.HasSuffix(filename, ".sql") {
			return nil
		}

		// Check if it's an up or down migration
		isDown := strings.Contains(filename, "_down.sql")
		if down && !isDown {
			return nil
		}
		if !down && isDown {
			return nil
		}

		// Extract migration name
		name := strings.TrimSuffix(filename, ".sql")
		if isDown {
			name = strings.TrimSuffix(name, "_down")
		}

		migrations = append(migrations, Migration{
			Name: name,
			Path: path,
			Type: down ? "down" : "up",
		})

		return nil
	})

	return migrations, err
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
} 