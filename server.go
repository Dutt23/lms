package main

import (
	"context"
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dutt23/lms/api"
	cache "github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/middleware"
	"github.com/dutt23/lms/pkg/connectors"
	service "github.com/dutt23/lms/services"
	"github.com/dutt23/lms/token"
	"github.com/dutt23/lms/workers"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
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

type routerOpts struct {
	bookCache   cache.BookCache
	bookService service.BookService

	memberCache   cache.MemberCache
	memberService service.MemberService

	loanService service.LoanService

	analyticsService service.AnalyticsService

	taskDistributor workers.TaskDistributor
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
	// Init cache
	bookCache := cache.NewBookCache(server.Cache)
	memberCache := cache.NewMemberCache(server.Cache)

	// Init Service
	bookservice := service.NewBookService(server.DB, bookCache)
	memberService := service.NewMemberService(server.DB, memberCache)
	loanService := service.NewLoanService(server.DB)
	analyticsService := service.NewAnalyticsService(bookCache, memberCache)

	redisOpts := asynq.RedisClientOpt{
		Addr: "0.0.0.0:6379",
	}
	// Queue
	taskDistributor := workers.NewRedisTaskDistributor(redisOpts)

	opts := &routerOpts{
		bookCache,
		bookservice,
		memberCache,
		memberService,

		loanService,

		analyticsService,

		taskDistributor,
	}
	// Add routes
	server.setupRouter(opts)
	return server, nil
}

func (s *Server) AllConnectors() {
	sql := connectors.NewSqliteConnector(&s.config.DbConfig)
	s.DB = sql
	cache := connectors.NewCacheConnector(&s.config.CacheConfig)
	s.Cache = cache
}

func (server *Server) setupRouter(opts *routerOpts) {
	router := gin.Default()
	apiv1 := router.Group("/v1/")
	server.addBookRoutes(apiv1, opts)
	server.addMemberRoutes(apiv1, opts)
	server.addLoanRoutes(apiv1, opts)
	server.addAnalyticsRoutes(apiv1, opts)
	server.addAuthRoutes(apiv1, opts)
	server.E = router
}

func (server *Server) addBookRoutes(grp *gin.RouterGroup, opts *routerOpts) {
	bookHandler := api.NewBooksApi(server.config, server.DB, opts.bookCache, opts.bookService)
	grp.POST("/books", bookHandler.AddBook)
	grp.GET("/books", bookHandler.GetBooks)
	grp.GET("/books/:id", bookHandler.GetBook)
	grp.PUT("/books/:id", bookHandler.UpdateBook)
	grp.DELETE("/books/:id", bookHandler.DeleteBook)
}

func (server *Server) addMemberRoutes(grp *gin.RouterGroup, opts *routerOpts) {
	memberHandler := api.NewMembersApi(server.config, server.DB, opts.memberCache, opts.memberService)
	grp.POST("/members", memberHandler.AddMember)
	grp.GET("/members", memberHandler.GetMembers)
	grp.GET("/members/:id", memberHandler.GetMember)
	grp.PUT("/members/:id", memberHandler.UpdateMember)
	grp.DELETE("/members/:id", memberHandler.DeleteMember)
}

func (server *Server) addLoanRoutes(grp *gin.RouterGroup, opts *routerOpts) {
	loansHandler := api.NewLoansApi(server.config, server.DB, opts.bookService, opts.memberService, opts.loanService, opts.taskDistributor)
	grp.POST("/loans", loansHandler.AddLoan)
	grp.GET("/loans", loansHandler.GetLoans)
	grp.GET("/loans/:id", loansHandler.GetLoan)
	grp.PUT("/loans/:id", loansHandler.UpdateLoan)
	grp.DELETE("/loans/:id", loansHandler.DeleteLoan)
}

func (server *Server) addAnalyticsRoutes(grp *gin.RouterGroup, opts *routerOpts) {
	analyticsHandler := api.NewAnalyticsApi(server.config, opts.bookService, opts.memberService, opts.analyticsService)
	grp.GET("/analytics", analyticsHandler.GetAnalytics)
}

func (server *Server) addAuthRoutes(grp *gin.RouterGroup, opts *routerOpts) {
	authHandler := api.NewAuthApi(server.config, opts.memberService, server.tokenMaker)
	grp.POST("/login/user", authHandler.LoginUser)
	authRoutes := grp.Group("/").Use(middleware.AuthMiddleware(server.tokenMaker))
	authRoutes.POST("/auth/check", authHandler.LoginUser)
}
