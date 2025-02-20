package main

import (
	"context"
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dutt23/lms/api"
	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/pkg/connectors"
	service "github.com/dutt23/lms/services"
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
	bookFilter *bloom.BloomFilter
}

func NewServer(config *config.AppConfig) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w", err)
	}

	bookFilter := bloom.NewWithEstimates(1000000, 0.01)
	server := &Server{tokenMaker: tokenMaker, config: config, bookFilter: bookFilter}

	// Init storages
	server.AllConnectors()
	// Add routes
	server.setupRouter()
	return server, nil
}

func (s *Server) AllConnectors() {
	sql := connectors.NewSqliteConnector(&s.config.DbConfig)
	s.DB = sql
	cache := connectors.NewCacheConnector(&s.config.CacheConfig)
	s.Cache = cache
}

func (server *Server) setupRouter() {
	router := gin.Default()
	apiv1 := router.Group("/v1/")
	server.addBookRoutes(apiv1)
	// authRoutes := router.Group("/").Use(middleware.AuthMiddleware(server.tokenMaker))
	server.E = router
}

func (server *Server) addBookRoutes(grp *gin.RouterGroup) {
	bookHandler := api.NewBooksApi(server.config, server.DB, server.bookFilter, service.NewBookCacheService(server.Cache))
	grp.POST("/book", bookHandler.AddBook)
	grp.GET("/books", bookHandler.GetBooks)
	grp.GET("/book/:id", bookHandler.GetBook)
	grp.PUT("/book/:id", bookHandler.UpdateBook)
}
