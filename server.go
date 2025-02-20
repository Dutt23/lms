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
	config       *config.AppConfig
	DB           connectors.SqliteConnector
	Cache        connectors.CacheConnector
	Closeable    []func(context.Context) error
	E            *gin.Engine
	tokenMaker   token.Maker
	bookFilter   *bloom.BloomFilter
	memberFilter *bloom.BloomFilter
}

func NewServer(config *config.AppConfig) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w", err)
	}

	bookFilter := bloom.NewWithEstimates(1000000, 0.01)
	memberFilter := bloom.NewWithEstimates(1000000, 0.01)
	server := &Server{tokenMaker: tokenMaker, config: config, bookFilter: bookFilter, memberFilter: memberFilter}

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
	server.addMemberRoutes(apiv1)
	// authRoutes := router.Group("/").Use(middleware.AuthMiddleware(server.tokenMaker))
	server.E = router
}

func (server *Server) addBookRoutes(grp *gin.RouterGroup) {
	bookHandler := api.NewBooksApi(server.config, server.DB, server.bookFilter, service.NewBookCacheService(server.Cache))
	grp.POST("/books", bookHandler.AddBook)
	grp.GET("/books", bookHandler.GetBooks)
	grp.GET("/books/:id", bookHandler.GetBook)
	grp.PUT("/books/:id", bookHandler.UpdateBook)
	grp.DELETE("/books/:id", bookHandler.DeleteBook)
}

func (server *Server) addMemberRoutes(grp *gin.RouterGroup) {
	memberHandler := api.NewMembersApi(server.config, server.DB, server.memberFilter, service.NewMemberCacheService(server.Cache))
	grp.POST("/members", memberHandler.AddMember)
	grp.GET("/members", memberHandler.GetMembers)
	grp.GET("/members/:id", memberHandler.GetMember)
	grp.PUT("/members/:id", memberHandler.UpdateMember)
	grp.DELETE("/member/:id", memberHandler.DeleteMember)
}
