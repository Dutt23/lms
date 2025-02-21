package api

import (
	"errors"
	"net/http"

	"github.com/dutt23/lms/config"
	service "github.com/dutt23/lms/services"
	"github.com/gin-gonic/gin"
)

type analyticsApi struct {
	config           *config.AppConfig
	bookService      service.BookService
	analyticsService service.AnalyticsService
}

func NewAnalyticsApi(config *config.AppConfig, bookService service.BookService, analyticsService service.AnalyticsService) *analyticsApi {
	return &analyticsApi{
		config,
		bookService,
		analyticsService,
	}
}

type getAnalyticsRequestBody struct {
}

type getAnalyticsResponseBody struct {
}

func (api *analyticsApi) GetAnalytics(ctx *gin.Context) {
	books, err := api.bookService.GetBooks(ctx, 0, 10)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("unable to get analytics")))
		return
	}

	bookIds := make([]uint64, len(books))

	for idx, book := range books {
		bookIds[idx] = book.Id
	}
	resp, err := api.analyticsService.GetBookListAnalytics(ctx, bookIds)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
