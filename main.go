package main

import (
	"context"

	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/pkg/connectors"
	"github.com/gin-gonic/gin"
)


type AppRunner struct {
	E         *gin.Engine
	Cfg       *config.AppConfig
	Sqlite    connectors.SqliteConnector
	Cache     connectors.CacheConnector
	Closeable []func(context.Context) error
}

func main() {}