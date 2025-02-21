package service

import (
	"context"

	"github.com/dutt23/lms/cache"
)

type analyticsService struct {
	cache cache.BookCache
}

func NewAnalyticsService(cache cache.BookCache) AnalyticsService {
	return &analyticsService{cache}
}

func (service *analyticsService) GetBookListAnalytics(ctx context.Context, bookIds []uint64) (*cache.BookAnalytics, error) {
	return service.cache.GetBookAnalytics(ctx, bookIds)
}
