package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"github.com/gin-gonic/gin"
)

type BookAddRequestBody struct {
  Title string `json:"title" binding:"required,alpha,gt=1"`
  Author string `json:"author" binding:"required,alpha,gt=1"`
  PublishedDate time.Time `json:"published_date" binding:"required,gt"`
  Isbn string `json:"isbn" binding:"required,alpha,gt=1"`
  NumberOfPages uint64 `json:"number_of_pages" binding:"required,numeric,gt=1"`
  CoverURL string `json:"cover_url" binding:"gt=1"`
  Language string `json:"language" binding:"required,alpha,gt=1"`
  AvailableCopies uint64 `json:"available_copies" binding:"required,numeric,gt=1"`
}

type booksApi struct {
  config *config.AppConfig
  db connectors.SqliteConnector
  filter *bloom.BloomFilter
}

func NewBooksApi(config *config.AppConfig, db connectors.SqliteConnector, bookFilter *bloom.BloomFilter) *booksApi {
  return &booksApi{
    config: config,
    db: db,
    filter: bookFilter,
  }
}

func (api booksApi) AddBook(ctx *gin.Context) {
  var req BookAddRequestBody 

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
  api.filter.Add([]byte(book.Isbn))
  ctx.JSON(http.StatusOK, book)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}