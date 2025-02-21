package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dutt23/lms/config"
	_ "github.com/dutt23/lms/docs"
	"github.com/dutt23/lms/workers"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/exp/rand"
)

type AppRunner struct {
	server    *Server
	Closeable []func(context.Context) error
}

func main() {
	ctx := context.Background()
	rand.Seed(uint64(time.Now().UnixNano()))
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
	defer appRunner.close(ctx)

	runMigrations(cfg.MigrationUrl, cfg.DBSource)
	appRunner.server.E.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	go appRunner.startProcessors(cfg)

	appRunner.server.E.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
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
		fmt.Println("error while connecting to postgres.", err)
		return err
	}
	app.server.DB = app.server.DB
	app.Closeable = append(app.Closeable, app.server.DB.Disconnect)

	err = app.server.Cache.Connect(ctx)
	if err != nil {
		fmt.Println("error while connecting to cache.", err)
		return err
	} else {
		app.Closeable = append(app.Closeable, app.server.Cache.Disconnect)
	}

	return nil
}

func (app *AppRunner) close(ctx context.Context) {
	if len(app.Closeable) > 0 {
		fmt.Println("there are closeable references to closed")
		for _, closeable := range app.Closeable {
			err := closeable(ctx)
			if err != nil {
				fmt.Println("error while closing %v", err)
			}
		}
	}
}

func (app *AppRunner) startProcessors(config *config.AppConfig) {
	analyticsProcessor := workers.NewAnalyticsTaskProcessor(config, app.server.Cache)
	fmt.Println("starting analytics processor")

	if err := analyticsProcessor.Start(); err != nil {
		fmt.Println("Unable to start analytics processor ", err)
	}
}

func runMigrations(migrationURL, dbSource string) {
	fmt.Println(migrationURL)
	fmt.Println(dbSource)
	m, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Panicf("Cannot create new migrate instance %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Panicf("Cannot run migrate up on the instance %w", err)
	}

	fmt.Println("database migration successful")
}
