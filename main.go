package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dutt23/lms/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	runMigrations(cfg.MigrationUrl, cfg.DBSource)
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
		fmt.Errorf("error while connecting to postgres.", err)
		return err
	}
	app.Closeable = append(app.Closeable, app.server.DB.Disconnect)

	err = app.server.Cache.Connect(ctx)
	if err != nil {
		fmt.Errorf("error while connecting to cache.", err)
		return err
	} else {
		app.Closeable = append(app.Closeable, app.server.Cache.Disconnect)
	}

	return nil
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
