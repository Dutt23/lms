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

func CacheKey(c context.Context, funcName string, key ...string) string {
	var builder strings.Builder
	builder.WriteString("INTERNAL::")
	builder.WriteString(funcName)
	builder.WriteString("_")
	builder.WriteString(strings.Join(key[:], ","))
	return builder.String()
}
