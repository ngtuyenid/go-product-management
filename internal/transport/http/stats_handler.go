package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thanhnguyen/product-api/internal/business/usecase"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// StatsHandler handles HTTP requests for statistics
type StatsHandler struct {
	statsUseCase usecase.StatsUseCase
	logger       *logger.Logger
}

// NewStatsHandler creates a new StatsHandler
func NewStatsHandler(statsUseCase usecase.StatsUseCase, logger *logger.Logger) *StatsHandler {
	return &StatsHandler{
		statsUseCase: statsUseCase,
		logger:       logger,
	}
}

// GetStats returns all statistics
func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.statsUseCase.GetStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCategoryStats returns product counts by category
func (h *StatsHandler) GetCategoryStats(c *gin.Context) {
	stats, err := h.statsUseCase.GetCategoryStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get category stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": stats})
}

// GetWishlistStats returns wishlist counts by product
func (h *StatsHandler) GetWishlistStats(c *gin.Context) {
	stats, err := h.statsUseCase.GetWishlistStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get wishlist stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wishlist stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wishlist_stats": stats})
}

// GetTopProducts returns top products by reviews
func (h *StatsHandler) GetTopProducts(c *gin.Context) {
	topProducts, err := h.statsUseCase.GetTopProducts(c.Request.Context(), 5)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get top products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"top_products": topProducts})
}

// RefreshStats forces a refresh of the statistics
func (h *StatsHandler) RefreshStats(c *gin.Context) {
	if err := h.statsUseCase.RefreshStats(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to refresh stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Statistics refreshed successfully"})
}

// RegisterRoutes registers the statistics routes
func (h *StatsHandler) RegisterRoutes(router *gin.RouterGroup) {
	stats := router.Group("/stats")
	{
		stats.GET("", h.GetStats)
		stats.GET("/categories", h.GetCategoryStats)
		stats.GET("/wishlist", h.GetWishlistStats)
		stats.GET("/top-products", h.GetTopProducts)
		stats.POST("/refresh", h.RefreshStats)
	}
}
