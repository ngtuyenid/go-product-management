package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/thanhnguyen/product-api/internal/business/usecase"
	"github.com/thanhnguyen/product-api/internal/config"
	"github.com/thanhnguyen/product-api/internal/transport/http/middleware"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	config         *config.Config
	logger         *logger.Logger
	authMiddleware *middleware.JWTAuthMiddleware
	rateLimiter    *middleware.IPRateLimiter
	errorHandler   *middleware.ErrorHandler
	productHandler *ProductHandler
	statsHandler   *StatsHandler
}

// NewServer creates a new HTTP server
func NewServer(
	config *config.Config,
	logger *logger.Logger,
	productUseCase usecase.ProductUseCase,
	statsUseCase usecase.StatsUseCase,
) *Server {
	// Set Gin mode
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Create server
	server := &Server{
		router: router,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Server.Port),
			Handler:      router,
			ReadTimeout:  config.Server.ReadTimeout,
			WriteTimeout: config.Server.WriteTimeout,
			IdleTimeout:  config.Server.IdleTimeout,
		},
		config: config,
		logger: logger,
	}

	// Initialize error handler
	server.errorHandler = middleware.NewErrorHandler(logger)
	router.Use(server.errorHandler.HandleErrors())
	router.NoRoute(server.errorHandler.NotFoundHandler())
	router.NoMethod(server.errorHandler.MethodNotAllowedHandler())

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     config.CORS.AllowOrigins,
		AllowMethods:     config.CORS.AllowMethods,
		AllowHeaders:     config.CORS.AllowHeaders,
		ExposeHeaders:    config.CORS.ExposeHeaders,
		AllowCredentials: config.CORS.AllowCredentials,
		MaxAge:           time.Duration(config.CORS.MaxAge) * time.Second,
	}
	router.Use(cors.New(corsConfig))

	// Initialize middleware
	server.authMiddleware = middleware.NewJWTAuthMiddleware(
		config.JWT.Secret,
		logger,
		time.Duration(config.JWT.ExpiryMinutes)*time.Minute,
	)

	// Initialize rate limiter
	server.rateLimiter = middleware.NewIPRateLimiter(
		config.RateLimit.Rate,
		config.RateLimit.Burst,
		logger,
	)
	server.rateLimiter.CleanupTask(
		time.Duration(config.RateLimit.CleanupIntervalMinutes)*time.Minute,
		time.Duration(config.RateLimit.ExpiryDurationMinutes)*time.Minute,
	)
	router.Use(server.rateLimiter.RateLimitMiddleware())

	// Setup middleware
	router.Use(gin.Logger())
	router.Use(server.requestLogger())

	// Setup handlers
	server.productHandler = NewProductHandler(productUseCase, logger)
	server.statsHandler = NewStatsHandler(statsUseCase, logger)

	// Register routes
	server.registerRoutes()

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Infof("Starting HTTP server on port %d", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// registerRoutes registers all HTTP routes
func (s *Server) registerRoutes() {
	// Public routes
	s.router.GET("/health", s.healthCheck)

	// Auth routes can be added here when needed

	// Protected API routes requiring authentication
	protectedAPI := s.router.Group("/api/v1")
	protectedAPI.Use(s.authMiddleware.Authenticate())
	{
		// Products
		s.productHandler.RegisterRoutes(protectedAPI)

		// Stats - require admin role
		statsRoutes := protectedAPI.Group("/stats")
		statsRoutes.Use(s.authMiddleware.AuthorizeRole("admin"))
		s.statsHandler.RegisterRoutes(protectedAPI)
	}
}

// healthCheck handles the health check endpoint
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// requestLogger logs request information
func (s *Server) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Log request details
		s.logger.WithFields(logger.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"duration": duration.String(),
			"ip":       c.ClientIP(),
		}).Info("Request processed")
	}
}
