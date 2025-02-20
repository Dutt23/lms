package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dutt23/lms/config"
	"github.com/gin-gonic/gin"
)

type AppRunner struct {
	server *Server
  Closeable []func(context.Context) error
}

func main() {
	ctx := context.Background()

	appRunner := AppRunner{}
	// resolving configuration
	cfg, err := appRunner.ResolveConfig()
	if err != nil {
		panic(err)
	}
	s, err := NewServer(cfg)
	if err != nil {
		panic(err)
	}
	appRunner.server = s
  appRunner.Init(ctx)
}

func (app *AppRunner) ResolveConfig() (*config.AppConfig, error) {
	vConfig, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Unable to parse viper config to application configuration : %v", err)
		return nil, err
	}

	cfg, err := config.GetApplicationConfig(vConfig)
	if err != nil {
		log.Fatalf("Unable to parse viper config to application configuration : %v", err)
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	// debug mode of gin when runing log in debug mode.
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	}
	return cfg, nil
}

func (app *AppRunner) Init(ctx context.Context) error {
	err := app.server.DB.Connect(ctx)
	if err != nil {
		fmt.Errorf("error while connecting to postgres.", err)
		return err
	}
	app.Closeable = append(app.Closeable, app.server.DB.Disconnect)
	return nil
}
