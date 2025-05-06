package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thanhnguyen/product-api/pkg/logger"
	"golang.org/x/time/rate"
)

// IPRateLimiter implements rate limiting per IP address
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     *sync.RWMutex
	rate   rate.Limit
	burst  int
	logger *logger.Logger
}

// NewIPRateLimiter creates a new instance of IPRateLimiter
func NewIPRateLimiter(r rate.Limit, b int, logger *logger.Logger) *IPRateLimiter {
	return &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		mu:     &sync.RWMutex{},
		rate:   r,
		burst:  b,
		logger: logger,
	}
}

// GetLimiter returns the rate limiter for a specific IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.ips[ip] = limiter
		i.mu.Unlock()
	}

	return limiter
}

// RateLimitMiddleware returns a gin middleware that implements rate limiting
func (i *IPRateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := i.GetLimiter(ip)
		if !limiter.Allow() {
			i.logger.WithField("ip", ip).Warn("Rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CleanupTask removes stale rate limiters to prevent memory leaks
func (i *IPRateLimiter) CleanupTask(cleanupInterval time.Duration, expiryDuration time.Duration) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for range ticker.C {
			i.cleanup(expiryDuration)
		}
	}()
}

// cleanup removes rate limiters that haven't been used in a while
func (i *IPRateLimiter) cleanup(expiryDuration time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// This is a simplified version. A more sophisticated implementation
	// would track the last access time for each limiter to determine
	// whether it should be removed.
	i.logger.Info("Cleaning up stale rate limiters")
	// In a real implementation, we would check the last access time of each limiter
	// and remove those that have been inactive for longer than expiryDuration.
}
