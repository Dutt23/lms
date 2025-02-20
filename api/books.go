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
  Title string `json:"title" binding:"required,alpha,gt=1"`
  Author string `json:"author" binding:"required,alpha,gt=1"`
  PublishedDate time.Time `json:"published_date" binding:"required,gt"`
  Isbn string `json:"isbn" binding:"required,alphanumeric,gt=1"`
  NumberOfPages uint64 `json:"number_of_pages" binding:"required,numeric,gt=1"`
  CoverURL string `json:"cover_url" binding:"gt=1"`
  Language string `json:"language" binding:"required,alpha,gt=1"`
  AvailableCopies uint64 `json:"available_copies" binding:"required,numeric,gt=1"`
}

type getBookRequestBody struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type booksApi struct {
  config *config.AppConfig
  db connectors.SqliteConnector
  filter *bloom.BloomFilter
  cache service.BookCacheService
}

func NewBooksApi(config *config.AppConfig, db connectors.SqliteConnector, bookFilter *bloom.BloomFilter, cache service.BookCacheService) *booksApi {
  return &booksApi{
    config: config,
    db: db,
    filter: bookFilter,
    cache: cache,
  }
}

func (api *booksApi) AddBook(ctx *gin.Context) {
  var req addBookRequestBody 

  	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

  book := &model.Book{
    Title: req.Title,
    Author: req.Author,
    PublishedDate: model.TimeWrapper(req.PublishedDate),
    Isbn: req.Isbn,
    NumberOfPages: req.NumberOfPages,
    CoverImage: req.CoverURL,
    Language: req.Language,
    AvailableCopies: req.AvailableCopies,
  }

  //TODO: Add retry logic here
  err := api.db.DB(ctx).Create(book).Error
  if err != nil {
    fmt.Errorf("error occurred %w", err)
    ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Unable to add book to library")))
  }
  
  go api.postProcessAddingBook(book)
  ctx.JSON(http.StatusOK, book)
}


func (api *booksApi) postProcessAddingBook(book *model.Book) {
  ctx := context.Background()
  api.filter.Add([]byte(book.Isbn))
  if err := api.cache.StoreBookMetaInCache(ctx, book); err != nil {
    fmt.Printf("Error occurred while adding book to cache %w", err)
  }
}


func (api *booksApi) GetBook(ctx *gin.Context) {
  var req getBookRequestBody

  if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}