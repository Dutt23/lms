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
	"gorm.io/gorm/clause"
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
	ID int64 `uri:"id" binding:"required,min=1"`
	addBookRequestBody
}

type getBookRequestBody struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

type deleteBookRequestBody struct {
	getBookRequestBody
}

type criteria struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Logic string `json:"logic"`
}

type getBooksRequestBody struct {
	LastId    int32       `json:"last_id"`
	Criterias []*criteria `json:"criterias"`
	PageSize  int32       `json:"page_size"`
}

type getBooksResponseBody struct {
	Books []*model.Book `json:"books"`
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

// AddBook godoc
//@Summary endpoint to create book
//@Description add a book
//@Tags book
//@Accept json
//@Produce json
//@Param book body addBookRequestBody true "Book data"
//@Success 200 {object} model.Book
//@Router /v1/book [post]
func (api *booksApi) AddBook(ctx *gin.Context) {
	var req addBookRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.IsIsbnUnique(ctx, req.Isbn) {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("duplicate isbn provided")))
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

// AddBook godoc
//@Summary endpoint to filter and get books
//@Description get a list of books
//@Tags book
//@Accept json
//@Produce json
//@Param bookListParams body getBooksRequestBody true "Book data"
//@Success 200 {object} []model.Book
//@Router /v1/book [get]
func (api *booksApi) GetBooks(ctx *gin.Context) {
	var req getBooksRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var books []*model.Book
	lastId := 0
	pageSize := 10

	if req.LastId > 0 {
		lastId = int(req.LastId)
	}

	if req.PageSize >= 100 || req.PageSize < 1 {
		pageSize = 10
	}

	qry := api.db.DB(ctx).Model(model.Book{}).Where("id > ?", lastId).Limit(pageSize)

	for _, ct := range req.Criterias {
		qry.Where(fmt.Sprintf("%s %s ?", ct.Key, ct.Logic), ct.Value)
	}

	tx := qry.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "published_date"},
		Desc:   true,
	}).Find(&books)

	if tx.Error != nil {
		fmt.Println("not able to find any books", tx.Error)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Unable to find any books")))
		return
	}

	ctx.JSON(http.StatusOK, getBooksResponseBody{Books: books})

}

// GetBook godoc
//@Summary endpoint to get book
//@Description get a book
//@Tags book
//@Produce json
//@param id path integer false "book id"
//@Success 200 {object} model.Book
//@Router /v1/book/:id [get]
func (api *booksApi) GetBook(ctx *gin.Context) {
	var req getBookRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := api.cache.GetBook(ctx, (req.ID))
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

	c := context.Background()
	go api.storeBookMeta(c, book)
	ctx.JSON(http.StatusOK, book)
}

// UpdateBook godoc
//@Summary endpoint to update book
//@Description update book data
//@Tags book
//@Produce json
//@Accept json
//@Param book body updateBookRequestBody true "Book data"
//@param id path integer false "book id"
//@Success 200 {object} model.Book
//@Router /v1/book/:id [put]
func (api *booksApi) UpdateBook(ctx *gin.Context) {
	var req updateBookRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.DoesBookExist(ctx, uint64(req.ID)) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("unable to locate book with Id %d", req.ID))))
		return 
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
	go api.postProcessAddingBook(book)
	ctx.JSON(http.StatusOK, book)
}

// DeleteBook godoc
//@Summary endpoint to delete book
//@Description delete a book
//@Tags book
//@param id path integer false "book id"
//@Success 200
//@Router /v1/book/:id [delete]
func (api *booksApi) DeleteBook(ctx *gin.Context) {
	var req deleteBookRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.DoesBookExist(ctx, uint64(req.ID)) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("unable to locate book with Id %d", req.ID))))
		return
	}

	if err := api.db.DB(ctx).Delete(&model.Book{}, req.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New(fmt.Sprintf("unable to delete book %d", req.ID))))
		return
	}

	go api.invalidateBookCache(req.ID)
	ctx.Status(http.StatusOK)
}

func (api *booksApi) postProcessAddingBook(book *model.Book) {
	ctx := context.Background()
	api.filter.Add([]byte(book.Isbn))
	api.storeBookMeta(ctx, book)
}

func (api *booksApi) storeBookMeta(ctx context.Context, book *model.Book) {
	if err := api.cache.StoreBookMetaInCache(ctx, book); err != nil {
		fmt.Printf("Error occurred while adding book to cache %w", err)
	}
}

func (api *booksApi) invalidateBookCache(bookId uint64) {
	ctx := context.Background()
	if err := api.cache.DeleteBook(ctx, bookId); err != nil {
		fmt.Println("Unable to remove entry from cache ", err)
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
