package http

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thanhnguyen/product-api/internal/business/usecase"
	"github.com/thanhnguyen/product-api/internal/transport/dto"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productUseCase usecase.ProductUseCase
	logger         *logger.Logger
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(productUseCase usecase.ProductUseCase, logger *logger.Logger) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
		logger:         logger,
	}
}

// CreateProduct handles product creation
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert DTO to entity
	product := req.ToEntity()

	// Call use case
	if err := h.productUseCase.CreateProduct(c.Request.Context(), product, req.CategoryIDs); err != nil {
		h.logger.WithError(err).Error("Failed to create product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Convert entity to response
	response := dto.FromEntity(*product)
	c.JSON(http.StatusCreated, response)
}

// GetProduct handles fetching a product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	// Parse ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Call use case
	product, err := h.productUseCase.GetProduct(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Convert entity to response
	response := dto.FromEntity(*product)
	c.JSON(http.StatusOK, response)
}

// ListProducts handles product listing with filtering and pagination
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values for pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// Convert DTO to filter
	filter := req.ToProductFilter()

	// Call use case
	products, totalItems, err := h.productUseCase.ListProducts(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products"})
		return
	}

	// Convert entities to response
	items := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		items = append(items, dto.FromEntity(p))
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalItems) / float64(req.PageSize)))

	// Build response
	response := dto.ProductListResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProduct handles product update
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Parse ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req dto.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert DTO to entity
	product := req.ToEntity()
	product.ID = uint(id)

	// Call use case
	if err := h.productUseCase.UpdateProduct(c.Request.Context(), product, req.CategoryIDs); err != nil {
		h.logger.WithError(err).Error("Failed to update product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Get updated product
	updatedProduct, err := h.productUseCase.GetProduct(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get updated product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated product"})
		return
	}

	// Convert entity to response
	response := dto.FromEntity(*updatedProduct)
	c.JSON(http.StatusOK, response)
}

// DeleteProduct handles product deletion
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Parse ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Call use case
	if err := h.productUseCase.DeleteProduct(c.Request.Context(), uint(id)); err != nil {
		h.logger.WithError(err).Error("Failed to delete product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *ProductHandler) SearchProductsByDescription(c *gin.Context) {
	desc := c.Query("query")
	if desc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query parameter"})
		return
	}
	products, err := h.productUseCase.SearchProductsByDescription(c.Request.Context(), desc)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}
	// TODO Convert to response DTO if needed
	c.JSON(http.StatusOK, products)
}

// RegisterRoutes registers the product routes
func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		products.POST("", h.CreateProduct)
		products.GET("", h.ListProducts)
		products.GET("/:id", h.GetProduct)
		products.PUT("/:id", h.UpdateProduct)
		products.DELETE("/:id", h.DeleteProduct)
		products.GET("/search", h.SearchProductsByDescription)
	}
}
