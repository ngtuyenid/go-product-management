package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/thanhnguyen/product-api/internal/business/entity"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// JWTAuthMiddleware provides JWT authentication functionality
type JWTAuthMiddleware struct {
	secretKey     []byte
	logger        *logger.Logger
	tokenDuration time.Duration
}

// JWTClaims represents the claims in a JWT
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTAuthMiddleware creates a new JWTAuthMiddleware
func NewJWTAuthMiddleware(secretKey string, logger *logger.Logger, tokenDuration time.Duration) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		secretKey:     []byte(secretKey),
		logger:        logger,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a new JWT token for a user
func (m *JWTAuthMiddleware) GenerateToken(user *entity.User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Authenticate validates the JWT token and sets the user in the context
func (m *JWTAuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the Authorization header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secretKey, nil
		})

		if err != nil {
			m.logger.WithError(err).Error("Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// AuthorizeRole checks if the user has the required role
func (m *JWTAuthMiddleware) AuthorizeRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "User not authorized for this action"})
		c.Abort()
	}
}

// RefreshToken refreshes an existing valid token
func (m *JWTAuthMiddleware) RefreshToken(c *gin.Context) {
	// Get the user information from the context (set by Authenticate middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	email, _ := c.Get("email")
	role, _ := c.Get("role")

	// Create a user entity from the context data
	user := &entity.User{
		ID:    userID.(uint),
		Email: email.(string),
		Role:  role.(string),
	}

	// Generate a new token
	token, err := m.GenerateToken(user)
	if err != nil {
		m.logger.WithError(err).Error("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"expires": time.Now().Add(m.tokenDuration),
	})
}
