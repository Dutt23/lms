package service

import (
	"context"
	"strings"
)

func CacheKey(c context.Context, funcName string, key ...string) string {
	var builder strings.Builder
	builder.WriteString("INTERNAL::")
	builder.WriteString(funcName)
	builder.WriteString("_")
	builder.WriteString(strings.Join(key[:], ","))
	return builder.String()
}