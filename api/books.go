package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	service "github.com/dutt23/lms/services"
	"github.com/gin-gonic/gin"
)

type addBookRequestBody struct {
	Title           string    `json:"title" binding:"required,gt=1"`
	Author          string    `json:"author" binding:"required,gt=1"`
	PublishedDate   time.Time `json:"published_date" binding:"required" time_format:"2006-01-02"`
	Isbn            string    `json:"isbn" binding:"required,alphanum,gt=1"`
	NumberOfPages   uint64    `json:"number_of_pages" binding:"required,numeric,gt=1"`
	CoverURL        string    `json:"cover_url" binding:"gt=1"`
	Language        string    `json:"language" binding:"required,alpha,gt=1"`
	AvailableCopies uint64    `json:"available_copies" binding:"required,numeric,gt=1"`
}

type updateBookRequestBody struct {
	ID              int64     `uri:"id" binding:"required,min=1"`
	addBookRequestBody
}

type getBookRequestBody struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

type booksApi struct {
	config *config.AppConfig
	db     connectors.SqliteConnector
	filter *bloom.BloomFilter
	cache  service.BookCacheService
}

func NewBooksApi(config *config.AppConfig, db connectors.SqliteConnector, bookFilter *bloom.BloomFilter, cache service.BookCacheService) *booksApi {
	return &booksApi{
		config: config,
		db:     db,
		filter: bookFilter,
		cache:  cache,
	}
}

func (api *booksApi) AddBook(ctx *gin.Context) {
	var req addBookRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	book := &model.Book{
		Title:           req.Title,
		Author:          req.Author,
		PublishedDate:   model.TimeWrapper(req.PublishedDate),
		Isbn:            req.Isbn,
		NumberOfPages:   req.NumberOfPages,
		CoverImage:      req.CoverURL,
		Language:        req.Language,
		AvailableCopies: req.AvailableCopies,
	}

	//TODO: Add retry logic here
	err := api.db.DB(ctx).Create(book).Error
	if err != nil {
		fmt.Errorf("error occurred %w", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Unable to add book to library")))
		return
	}

	go api.postProcessAddingBook(book)
	ctx.JSON(http.StatusCreated, book)
}

func (api *booksApi) GetBook(ctx *gin.Context) {
	var req getBookRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.DoesBookExist(ctx, uint64(req.ID)) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("unable to locate book with Id %d", req.ID))))
		return
	}

	res := api.cache.GetBook(ctx, (req.ID)); 
	if res != nil {
		ctx.JSON(http.StatusOK, res)
		return 
	}

	db := api.db.DB(ctx)
	var book *model.Book
	if err := db.Last(&book, req.ID).Error; err != nil {
		fmt.Errorf("error : %w", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New(fmt.Sprintf("unable to locate book with Id %d", req.ID))))
		return
	}
	ctx.JSON(http.StatusOK, book)
}

func (api *booksApi) UpdateBook(ctx *gin.Context) {
	var req updateBookRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.DoesBookExist(ctx, uint64(req.ID)) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("unable to locate book with Id %d", req.ID))))
	}

	book := &model.Book{
		Audited: model.Audited{
			Id: uint64(req.ID),
		},
		Title:           req.Title,
		Author:          req.Author,
		PublishedDate:   model.TimeWrapper(req.PublishedDate),
		Isbn:            req.Isbn,
		NumberOfPages:   req.NumberOfPages,
		CoverImage:      req.CoverURL,
		Language:        req.Language,
		AvailableCopies: req.AvailableCopies,
	}

	if err := api.db.DB(ctx).Save(book).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	ctx.JSON(http.StatusOK, book)
}

func (api *booksApi) postProcessAddingBook(book *model.Book) {
	ctx := context.Background()
	api.filter.Add([]byte(book.Isbn))
	if err := api.cache.StoreBookMetaInCache(ctx, book); err != nil {
		fmt.Printf("Error occurred while adding book to cache %w", err)
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
