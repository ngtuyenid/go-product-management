package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/thanhnguyen/product-api/internal/business/entity"
	"github.com/thanhnguyen/product-api/internal/storage"
	"github.com/thanhnguyen/product-api/internal/storage/cache"
	transportHttp "github.com/thanhnguyen/product-api/internal/transport/http"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

// StatsUseCase defines the statistics business logic
type StatsUseCase interface {
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetCategoryStats(ctx context.Context) ([]entity.CategoryStat, error)
	GetWishlistStats(ctx context.Context) ([]entity.WishlistStat, error)
	GetTopProducts(ctx context.Context, limit int) ([]entity.TopProduct, error)
	RefreshStats(ctx context.Context) error
}

// statsUseCase implements StatsUseCase
type statsUseCase struct {
	productRepo    storage.ProductRepository
	categoryRepo   storage.CategoryRepository
	wishlistRepo   storage.WishlistRepository
	reviewRepo     storage.ReviewRepository
	cache          *cache.StatsCache
	logger         *logger.Logger
	refreshTimeout time.Duration
	lastRefresh    time.Time
	mutex          sync.RWMutex
	wsHub          *transportHttp.WebSocketHub
}

// NewStatsUseCase creates a new StatsUseCase
func NewStatsUseCase(
	productRepo storage.ProductRepository,
	categoryRepo storage.CategoryRepository,
	wishlistRepo storage.WishlistRepository,
	reviewRepo storage.ReviewRepository,
	cache *cache.StatsCache,
	logger *logger.Logger,
	refreshTimeout time.Duration,
	wsHub *transportHttp.WebSocketHub,
) StatsUseCase {
	// Create the use case
	uc := &statsUseCase{
		productRepo:    productRepo,
		categoryRepo:   categoryRepo,
		wishlistRepo:   wishlistRepo,
		reviewRepo:     reviewRepo,
		cache:          cache,
		logger:         logger,
		refreshTimeout: refreshTimeout,
		wsHub:          wsHub,
	}

	// Do an initial refresh
	go uc.RefreshStats(context.Background())

	// Start the background refresh goroutine
	go uc.startRefreshLoop()

	return uc
}

// startRefreshLoop periodically refreshes the statistics
func (uc *statsUseCase) startRefreshLoop() {
	ticker := time.NewTicker(uc.refreshTimeout)
	defer ticker.Stop()

	for range ticker.C {
		if err := uc.RefreshStats(context.Background()); err != nil {
			uc.logger.WithError(err).Error("Failed to refresh statistics")
		}
	}
}

// GetStats returns all statistics
func (uc *statsUseCase) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Check if stats need to be refreshed
	uc.mutex.RLock()
	needsRefresh := time.Since(uc.lastRefresh) > uc.refreshTimeout
	uc.mutex.RUnlock()

	if needsRefresh {
		if err := uc.RefreshStats(ctx); err != nil {
			return nil, err
		}
	}

	// Get all stats from cache
	return uc.cache.GetAll(), nil
}

// GetCategoryStats returns product counts by category
func (uc *statsUseCase) GetCategoryStats(ctx context.Context) ([]entity.CategoryStat, error) {
	// Get category counts from cache
	categoryCounts := uc.cache.GetCategoryCounts()

	// Check if we need to refresh
	if len(categoryCounts) == 0 {
		if err := uc.RefreshStats(ctx); err != nil {
			return nil, err
		}
		categoryCounts = uc.cache.GetCategoryCounts()
	}

	// Get all categories for names
	categories, err := uc.categoryRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Create the result with category names
	stats := make([]entity.CategoryStat, 0, len(categoryCounts))
	for id, count := range categoryCounts {
		categoryName := "Unknown"
		for _, cat := range categories {
			if cat.ID == id {
				categoryName = cat.Name
				break
			}
		}

		stats = append(stats, entity.CategoryStat{
			CategoryID:   id,
			CategoryName: categoryName,
			ProductCount: count,
		})
	}

	return stats, nil
}

// GetWishlistStats returns wishlist counts by product
func (uc *statsUseCase) GetWishlistStats(ctx context.Context) ([]entity.WishlistStat, error) {
	// Get wishlist counts from cache
	wishlistCounts := uc.cache.GetWishlistCounts()

	// Check if we need to refresh
	if len(wishlistCounts) == 0 {
		if err := uc.RefreshStats(ctx); err != nil {
			return nil, err
		}
		wishlistCounts = uc.cache.GetWishlistCounts()
	}

	// Create the result
	stats := make([]entity.WishlistStat, 0, len(wishlistCounts))

	// Use a waitgroup to fetch product details concurrently
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for id, count := range wishlistCounts {
		wg.Add(1)
		go func(id uint, count int) {
			defer wg.Done()

			// Get product details
			product, err := uc.productRepo.FindByID(ctx, id)
			if err != nil {
				uc.logger.WithError(err).Error("Failed to get product details for wishlist stats")
				return
			}

			if product != nil {
				stat := entity.WishlistStat{
					ProductID:     id,
					ProductName:   product.Name,
					WishlistCount: count,
				}

				mu.Lock()
				stats = append(stats, stat)
				mu.Unlock()
			}
		}(id, count)
	}

	wg.Wait()

	return stats, nil
}

// GetTopProducts returns the top products by review count
func (uc *statsUseCase) GetTopProducts(ctx context.Context, limit int) ([]entity.TopProduct, error) {
	// Check if we have cached top products
	if value, exists := uc.cache.Get("top_products"); exists {
		if topProducts, ok := value.([]entity.TopProduct); ok {
			return topProducts, nil
		}
	}

	// If not cached, refresh the stats
	if err := uc.RefreshStats(ctx); err != nil {
		return nil, err
	}

	// Try again from cache
	if value, exists := uc.cache.Get("top_products"); exists {
		if topProducts, ok := value.([]entity.TopProduct); ok {
			return topProducts, nil
		}
	}

	// If still not available, return empty slice
	return []entity.TopProduct{}, nil
}

// RefreshStats refreshes all statistics
func (uc *statsUseCase) RefreshStats(ctx context.Context) error {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	uc.logger.Info("Refreshing statistics")

	// Use waitgroup to parallelize stat collection
	var (
		wg                sync.WaitGroup
		productCount      int64
		userCount         int64
		reviewCount       int64
		avgRating         float64
		categoryCounts    map[uint]int
		wishlistCounts    map[uint]int
		topProducts       []entity.TopProduct
		productCountErr   error
		userCountErr      error
		reviewCountErr    error
		avgRatingErr      error
		categoryCountsErr error
		wishlistCountsErr error
		topProductsErr    error
	)

	// Get total product count
	wg.Add(1)
	go func() {
		defer wg.Done()
		var result []entity.Product
		var err error
		result, productCount, err = uc.productRepo.List(ctx, entity.ProductFilter{Page: 1, PageSize: 1})
		if err != nil {
			productCountErr = err
			uc.logger.WithError(err).Error("Failed to count products")
		}
		_ = result // Avoid unused variable warning
	}()

	// Get category counts
	wg.Add(1)
	go func() {
		defer wg.Done()

		// This would normally call a repository method, but for now we'll simulate
		// with a direct SQL query

		// TODO: Implement repository method for category counts
		categoryCounts = make(map[uint]int)
		categoryCountsErr = nil
	}()

	// Get wishlist counts
	wg.Add(1)
	go func() {
		defer wg.Done()

		// This would normally call a repository method, but for now we'll simulate
		// with a direct SQL query

		// TODO: Implement repository method for wishlist counts
		wishlistCounts = make(map[uint]int)
		wishlistCountsErr = nil
	}()

	// Get top products
	wg.Add(1)
	go func() {
		defer wg.Done()

		// This would normally call a repository method, but for now we'll simulate

		// TODO: Implement repository method for top products
		topProducts = make([]entity.TopProduct, 0)
		topProductsErr = nil
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Check for errors
	if productCountErr != nil {
		return productCountErr
	}
	if userCountErr != nil {
		return userCountErr
	}
	if reviewCountErr != nil {
		return reviewCountErr
	}
	if avgRatingErr != nil {
		return avgRatingErr
	}
	if categoryCountsErr != nil {
		return categoryCountsErr
	}
	if wishlistCountsErr != nil {
		return wishlistCountsErr
	}
	if topProductsErr != nil {
		return topProductsErr
	}

	// Update the cache
	uc.cache.Set("total_products", productCount)
	uc.cache.Set("total_users", userCount)
	uc.cache.Set("total_reviews", reviewCount)
	uc.cache.Set("average_rating", avgRating)
	uc.cache.Set("top_products", topProducts)
	uc.cache.SetCategoryCounts(categoryCounts)
	uc.cache.SetWishlistCounts(wishlistCounts)

	// Update last refresh time
	uc.lastRefresh = time.Now()

	uc.logger.Info("Statistics refreshed")

	// Broadcast stats update
	uc.wsHub.Broadcast([]byte(`{"event":"stats_update","data":...}`))

	return nil
}
