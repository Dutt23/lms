package main

import (
	"context"
	"fmt"

	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/pkg/connectors"
	"github.com/dutt23/lms/token"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *config.AppConfig
	DB         connectors.SqliteConnector
	Cache      connectors.CacheConnector
	Closeable  []func(context.Context) error
	E          *gin.Engine
	tokenMaker token.Maker
}

func NewServer(config *config.AppConfig) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w", err)
	}

	server := &Server{tokenMaker: tokenMaker, config: config}

	// Init storages
	server.AllConnectors()
	// Add routes
	server.setupRouter()
	return server, nil
}

func (s *Server) AllConnectors() {
	sql := connectors.NewSqliteConnector(&s.config.DbConfig)
	s.DB = sql
	// redis := connectors.NewRedisConnector(&g.Cfg.RedisConfig, g.Logger)
	// g.Redis = redis
}

func (server *Server) setupRouter() {
	router := gin.Default()
	server.E = router
}
