package cache

import (
	"context"
	"strings"
)

const (
	BOOK_ISBN_FILTER    = "books:isbn"
	MEMBER_EMAIL_FILTER = "members:email"
)

type Cache struct{}

type BookAnalytics struct {
	Analytics map[string]*BookAnalytic `json:"book_analytics"`
}

type BookAnalytic struct {
	Analytics []*BookFreq `json:"analytics"`
}

type BookFreq struct {
	Month string `json:"month"`
	Count uint64 `json:"count"`
}

func CacheKey(c context.Context, funcName string, key ...string) string {
	var builder strings.Builder
	builder.WriteString("INTERNAL::")
	builder.WriteString(funcName)
	builder.WriteString("_")
	builder.WriteString(strings.Join(key[:], ","))
	return builder.String()
}
