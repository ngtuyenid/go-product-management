package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

// Config holds all configuration for the application
type Config struct {
	Environment   string
	Server        ServerConfig
	Database      DatabaseConfig
	JWT           JWTConfig
	CORS          CORSConfig
	RateLimit     RateLimitConfig
	Logger        LoggerConfig
	Elasticsearch ElasticsearchConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
	MaxConns int
	MinConns int
	Timeout  time.Duration
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret        string
	ExpiryMinutes int
}

// CORSConfig holds CORS-specific configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Rate                   rate.Limit
	Burst                  int
	CleanupIntervalMinutes int
	ExpiryDurationMinutes  int
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string
	Format     string
	OutputPath string
}

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	URL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 10)) * time.Second,
			IdleTimeout:  time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "product_api"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 10),
			MinConns: getEnvAsInt("DB_MIN_CONNS", 2),
			Timeout:  time.Duration(getEnvAsInt("DB_TIMEOUT", 5)) * time.Second,
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryMinutes: getEnvAsInt("JWT_EXPIRY_MINUTES", 60),
		},
		CORS: CORSConfig{
			AllowOrigins:     getEnvAsSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
			AllowMethods:     getEnvAsSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowHeaders:     getEnvAsSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
			ExposeHeaders:    getEnvAsSlice("CORS_EXPOSE_HEADERS", []string{}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 300),
		},
		RateLimit: RateLimitConfig{
			Rate:                   rate.Limit(getEnvAsFloat("RATE_LIMIT_RATE", 10)),
			Burst:                  getEnvAsInt("RATE_LIMIT_BURST", 20),
			CleanupIntervalMinutes: getEnvAsInt("RATE_LIMIT_CLEANUP_INTERVAL", 5),
			ExpiryDurationMinutes:  getEnvAsInt("RATE_LIMIT_EXPIRY_DURATION", 60),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOGGER_LEVEL", "info"),
			Format:     getEnv("LOGGER_FORMAT", "json"),
			OutputPath: getEnv("LOGGER_OUTPUT_PATH", "stdout"),
		},
	}

	return config, nil
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.Username,
		c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

// Helper functions to get environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
