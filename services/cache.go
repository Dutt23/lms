package service

import (
	"context"
	"fmt"
	"strings"
)

func CacheKey(c context.Context, funcName string, key ...string) string {
	fmt.Println(key)
	var builder strings.Builder
	builder.WriteString("INTERNAL::")
	builder.WriteString(funcName)
	builder.WriteString("_")
	builder.WriteString(strings.Join(key[:], ","))
	return builder.String()
}
