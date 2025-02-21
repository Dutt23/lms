package api

import (
	"context"
	"net/http"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/config"
	service "github.com/dutt23/lms/services"
	"github.com/gin-gonic/gin"
)

type analyticsApi struct {
	config           *config.AppConfig
	bookService      service.BookService
	memberService    service.MemberService
	analyticsService service.AnalyticsService
}

func NewAnalyticsApi(config *config.AppConfig,
	bookService service.BookService,
	memberService service.MemberService,
	analyticsService service.AnalyticsService) *analyticsApi {
	return &analyticsApi{
		config,
		bookService,
		memberService,
		analyticsService,
	}
}

type getAnalyticsRequestBody struct {
}

type getAnalyticsResponseBody struct {
	BookAnalytics   *cache.BookAnalytics   `json:"book_month_analytics"`
	MemberAnalytics *cache.MemberAnalytics `json:"member_week_analytics"`
}


// GetAnalytics godoc
// @Summary endpoint to ten latest books and members
// @Description get analytics
// @Tags analytics
// @Produce json
// @Success 200 {object} getAnalyticsResponseBody
// @Router /v1/analytics [get]
func (api *analyticsApi) GetAnalytics(ctx *gin.Context) {
	bookResp, err := api.getBookAnalytics(ctx)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	resp := getAnalyticsResponseBody{
		BookAnalytics: bookResp,
	}

	memberResp, err := api.getMemberAnalytics(ctx)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp.MemberAnalytics = memberResp

	ctx.JSON(http.StatusOK, resp)
}

func (api *analyticsApi) getBookAnalytics(ctx context.Context) (*cache.BookAnalytics, error) {
	books, err := api.bookService.GetBooks(ctx, 0, 10)
	if err != nil {
		return nil, err
	}

	bookIds := make([]uint64, len(books))

	for idx, book := range books {
		bookIds[idx] = book.Id
	}
	bookResp, err := api.analyticsService.GetBookListAnalytics(ctx, bookIds)

	if err != nil {
		return nil, err
	}

	return bookResp, nil
}

func (api *analyticsApi) getMemberAnalytics(ctx context.Context) (*cache.MemberAnalytics, error) {
	members, err := api.memberService.GetMembers(ctx, 0, 10)
	if err != nil {
		return nil, err
	}

	memberIds := make([]uint64, len(members))

	for idx, book := range members {
		memberIds[idx] = book.Id
	}
	memberResp, err := api.analyticsService.GetMemberListAnalytics(ctx, memberIds)

	if err != nil {
		return nil, err
	}

	return memberResp, nil
}
