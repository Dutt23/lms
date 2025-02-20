package main

import (
	"context"
	"log"

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

func main() {
  	ctx := context.Background()

	appRunner := AppRunner{E: gin.New()}
	// resolving configuration
	err := appRunner.ResolveConfig()
}

func (app *AppRunner) ResolveConfig() error {
	vConfig, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Unable to parse viper config to application configuration : %v", err)
		return err
	}

	cfg, err := config.GetApplicationConfig(vConfig)
	if err != nil {
		log.Fatalf("Unable to parse viper config to application configuration : %v", err)
		return err
	}

	app.Cfg = cfg
	gin.SetMode(gin.ReleaseMode)
	// debug mode of gin when runing log in debug mode.
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	}
	return nil

}