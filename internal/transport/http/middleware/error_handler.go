package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ErrorHandler provides error handling middleware
type ErrorHandler struct {
	logger *logger.Logger
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler(logger *logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleErrors returns middleware that handles and logs errors
func (h *ErrorHandler) HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err

			// Log the error
			h.logger.WithField("path", c.Request.URL.Path).
				WithField("method", c.Request.Method).
				WithField("client_ip", c.ClientIP()).
				WithError(err).
				Error("Request error")

			// Check if response was already written
			if c.Writer.Written() {
				return
			}

			// Prepare error response
			status := http.StatusInternalServerError
			message := "Internal server error"
			errorMsg := err.Error()

			// Check if the error is already handled by other middleware
			if c.Writer.Status() != http.StatusOK {
				status = c.Writer.Status()
			}

			// Set appropriate message based on status code
			switch status {
			case http.StatusNotFound:
				message = "Resource not found"
			case http.StatusBadRequest:
				message = "Invalid request"
			case http.StatusUnauthorized:
				message = "Authentication required"
			case http.StatusForbidden:
				message = "Access denied"
			case http.StatusTooManyRequests:
				message = "Rate limit exceeded"
			}

			// Respond with JSON
			c.JSON(status, ErrorResponse{
				Status:  status,
				Message: message,
				Error:   errorMsg,
			})
		}
	}
}

// NotFoundHandler handles 404 errors
func (h *ErrorHandler) NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.logger.WithField("path", c.Request.URL.Path).
			WithField("method", c.Request.Method).
			WithField("client_ip", c.ClientIP()).
			Warn("Resource not found")

		c.JSON(http.StatusNotFound, ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Resource not found",
			Error:   "The requested URL was not found on the server",
		})
	}
}

// MethodNotAllowedHandler handles 405 errors
func (h *ErrorHandler) MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.logger.WithField("path", c.Request.URL.Path).
			WithField("method", c.Request.Method).
			WithField("client_ip", c.ClientIP()).
			Warn("Method not allowed")

		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
			Status:  http.StatusMethodNotAllowed,
			Message: "Method not allowed",
			Error:   "The method is not allowed for the requested URL",
		})
	}
}
