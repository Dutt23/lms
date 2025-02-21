package service

import (
	"context"

	"github.com/dutt23/lms/cache"
)

type analyticsService struct {
	cache       cache.BookCache
	memberCache cache.MemberCache
}

func NewAnalyticsService(cache cache.BookCache, memberCache cache.MemberCache) AnalyticsService {
	return &analyticsService{cache, memberCache}
}

func (service *analyticsService) GetBookListAnalytics(ctx context.Context, bookIds []uint64) (*cache.BookAnalytics, error) {
	return service.cache.GetBookAnalytics(ctx, bookIds)
}

func (service *analyticsService) GetMemberListAnalytics(ctx context.Context, memberIds []uint64) (*cache.MemberAnalytics, error) {
	return service.memberCache.GetMemberAnalytics(ctx, memberIds)
}
