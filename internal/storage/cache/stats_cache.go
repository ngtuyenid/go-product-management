package cache

import (
	"sync"
	"time"

	"github.com/thanhnguyen/product-api/pkg/logger"
)

// StatsCache provides caching for real-time statistics
type StatsCache struct {
	data           map[string]interface{}
	categoryCounts map[uint]int
	wishlistCounts map[uint]int
	mutex          sync.RWMutex
	lastRefreshed  time.Time
	logger         *logger.Logger
}

// NewStatsCache creates a new StatsCache
func NewStatsCache(logger *logger.Logger) *StatsCache {
	return &StatsCache{
		data:           make(map[string]interface{}),
		categoryCounts: make(map[uint]int),
		wishlistCounts: make(map[uint]int),
		mutex:          sync.RWMutex{},
		logger:         logger,
	}
}

// Set stores a value in the cache
func (c *StatsCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
	c.lastRefreshed = time.Now()
}

// Get retrieves a value from the cache
func (c *StatsCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

// GetAll returns all cached data
func (c *StatsCache) GetAll() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Create a copy of the data to avoid concurrent access issues
	result := make(map[string]interface{}, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}

	// Add metadata
	result["last_refreshed"] = c.lastRefreshed.Format(time.RFC3339)

	return result
}

// SetCategoryCounts sets the product counts by category
func (c *StatsCache) SetCategoryCounts(counts map[uint]int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a copy of the counts
	c.categoryCounts = make(map[uint]int, len(counts))
	for k, v := range counts {
		c.categoryCounts[k] = v
	}

	c.lastRefreshed = time.Now()
}

// GetCategoryCounts gets the product counts by category
func (c *StatsCache) GetCategoryCounts() map[uint]int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Create a copy of the counts
	result := make(map[uint]int, len(c.categoryCounts))
	for k, v := range c.categoryCounts {
		result[k] = v
	}

	return result
}

// SetWishlistCounts sets the wishlist counts by product
func (c *StatsCache) SetWishlistCounts(counts map[uint]int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a copy of the counts
	c.wishlistCounts = make(map[uint]int, len(counts))
	for k, v := range counts {
		c.wishlistCounts[k] = v
	}

	c.lastRefreshed = time.Now()
}

// GetWishlistCounts gets the wishlist counts by product
func (c *StatsCache) GetWishlistCounts() map[uint]int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Create a copy of the counts
	result := make(map[uint]int, len(c.wishlistCounts))
	for k, v := range c.wishlistCounts {
		result[k] = v
	}

	return result
}

// Clear clears all cached data
func (c *StatsCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]interface{})
	c.categoryCounts = make(map[uint]int)
	c.wishlistCounts = make(map[uint]int)
	c.lastRefreshed = time.Now()
}

// GetLastRefreshed returns the time when the cache was last refreshed
func (c *StatsCache) GetLastRefreshed() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lastRefreshed
}
