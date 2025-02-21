package service

import (
	"context"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
)


type bookService struct {
	db connectors.SqliteConnector
  cache  cache.BookCache
}

func NewBookService(db connectors.SqliteConnector, cache  cache.BookCache) BookService {
  return &bookService{db, cache}
}

func (service *bookService) GetBook(ctx context.Context, bookId uint64) (*model.Book, error) {
  res := service.cache.GetBook(ctx, bookId)
	if res != nil {
		return res, nil
	}

	db := service.db.DB(ctx)
	var book *model.Book
	if err := db.Last(&book, bookId).Error; err != nil {
		return nil, err
	}

  return book, nil
}