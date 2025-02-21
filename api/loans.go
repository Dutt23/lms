package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	service "github.com/dutt23/lms/services"
	"github.com/dutt23/lms/workers"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type loansApi struct {
	config          *config.AppConfig
	db              connectors.SqliteConnector
	bookService     service.BookService
	memberService   service.MemberService
	loanService     service.LoanService
	taskDistributor workers.TaskDistributor
}

func NewLoansApi(config *config.AppConfig,
	db connectors.SqliteConnector,
	bookService service.BookService,
	memberService service.MemberService,
	loanService service.LoanService,
	taskDistributor workers.TaskDistributor,
) *loansApi {
	return &loansApi{
		config:          config,
		db:              db,
		bookService:     bookService,
		memberService:   memberService,
		loanService:     loanService,
		taskDistributor: taskDistributor,
	}
}

type addLoanRequestBody struct {
	MemberId   uint64     `json:"member_id" binding:"required,numeric"`
	BookId     uint64     `json:"book_id" binding:"required,numeric"`
	ReturnDate *time.Time `json:"return_date"`
}

type getLoanRequestBody struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateLoanRequestBody struct {
	getLoanRequestBody
}

type deleteLoanRequestBody struct {
	getLoanRequestBody
}

type getLoansRequestBody struct {
	LastId   uint64 `json:"last_id"`
	PageSize int32  `json:"page_size"`
}

// AddLoadn godoc
// @Summary endpoint to loan a book
// @Description loan a book
// @Tags loan
// @Accept json
// @Produce json
// @Param book body addLoanRequestBody true "Loan data"
// @Success 200 {object} model.Loan
// @Router /v1/loans [post]
func (api *loansApi) AddLoan(ctx *gin.Context) {
	var req addLoanRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	book, err := api.bookService.GetBook(ctx, req.BookId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//TODO: Add cache check here

	if book.AvailableCopies <= 0 {
		ctx.JSON(http.StatusBadRequest, errors.New("not enough copies available of this book"))
		return
	}

	if err := api.bookService.ChangeAvailableCopies(ctx, req.BookId, -1); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	loan, err := api.loanService.SaveLoan(ctx, req.MemberId, req.BookId, req.ReturnDate)
	if err != nil {
		fmt.Println("unable to loan book , ", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	go api.startAnalytics(loan, book)
	ctx.JSON(http.StatusCreated, loan)
}

// GetLoan godoc
// @Summary endpoint to get loan
// @Description get a loan
// @Tags loan
// @Accept json
// @Produce json
// @param id path integer false "loan id"
// @Success 200 {object} []model.Loan
// @Router /v1/loans/:id [get]
func (api *loansApi) GetLoan(ctx *gin.Context) {
	var req getLoanRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	loan, err := api.loanService.GetLoan(ctx, uint64(req.ID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, loan)
}

// UpdateLoan godoc
// @Summary endpoint to mark a loan as completed
// @Description mark loan as completed
// @Tags loan
// @Accept json
// @Produce json
// @param id path integer false "loan id"
// @Router /v1/loans/:id [put]
func (api *loansApi) UpdateLoan(ctx *gin.Context) {
	var req updateLoanRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	loan, err := api.loanService.GetLoan(ctx, uint64(req.ID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = api.loanService.CompleteLoan(ctx, uint64(req.ID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = api.bookService.ChangeAvailableCopies(ctx, loan.BookId, 1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.Status(http.StatusOK)
}

// GetLoans godoc
// @Summary endpoint to filter and get loans
// @Description get a list of loans
// @Tags loan
// @Accept json
// @Produce json
// @Param loanListParams body getLoansRequestBody true "Loan data"
// @Success 200 {object} []model.Loan
// @Router /v1/loans [get]
func (api *loansApi) GetLoans(ctx *gin.Context) {
	var req getLoansRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	lastId := 0
	pageSize := 10

	if req.LastId > 0 {
		lastId = int(req.LastId)
	}

	if req.PageSize >= 100 || req.PageSize < 1 {
		pageSize = 10
	}

	loans, err := api.loanService.GetLoans(ctx, uint64(lastId), pageSize)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, loans)
}

// DeleteLoan godoc
// @Summary endpoint to delete loan
// @Description delete a loan
// @Tags loan
// @param id path integer false "loan id"
// @Success 200
// @Router /v1/loans/:id [delete]
func (api *loansApi) DeleteLoan(ctx *gin.Context) {
	var req deleteLoanRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	loan, err := api.loanService.GetLoan(ctx, uint64(req.ID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = api.loanService.DeleteLoan(ctx, uint64(req.ID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = api.bookService.ChangeAvailableCopies(ctx, loan.BookId, 1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.Status(http.StatusOK)
}

func (api *loansApi) startAnalytics(loan *model.Loan, book *model.Book) {
	ctx := context.Background()
	member, err := api.memberService.GetMember(ctx, loan.MemberId)

	if err != nil {
		fmt.Println("unable to enque task ", member)
		return
	}
	payload := &workers.BookAnalyticsPayload{
		Book:   book,
		Loan:   loan,
		Member: member,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(workers.CriticalQueue),
	}
	err = api.taskDistributor.DistributeBooksAnalyticsPayload(ctx, payload, opts...)
	if err != nil {
		fmt.Println("unable to queue analytics task ", err)
	}
}
