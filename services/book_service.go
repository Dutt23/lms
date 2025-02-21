package service

import (
	"context"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
)

type bookService struct {
	db    connectors.SqliteConnector
	cache cache.BookCache
}

func NewBookService(db connectors.SqliteConnector, cache cache.BookCache) BookService {
	return &bookService{db, cache}
}

func (service *bookService) GetBook(ctx context.Context, bookId uint64) (*model.Book, error) {
	res, err := service.cache.GetBook(ctx, bookId)
	if err == nil {
		return res, nil
	}

	db := service.db.DB(ctx)
	var book *model.Book
	if err := db.Last(&book, bookId).Error; err != nil {
		return nil, err
	}

	return book, nil
}

func (service *bookService) ChangeAvailableCopies(ctx context.Context, bookId uint64, count int64) error {
	db := service.db.DB(ctx)
	res, err := service.cache.GetBook(ctx, bookId)
	if res != nil && err == nil {
		res.AvailableCopies = res.AvailableCopies + count
		db.Save(res)
		service.cache.StoreBookMetaInCache(ctx, res)
		return nil
	}

	var book *model.Book
	if err := db.Last(&book, bookId).Error; err != nil {
		return err
	}

	book.AvailableCopies = book.AvailableCopies + count
	db.Save(res)
	service.cache.StoreBookMetaInCache(ctx, res)
	return nil
}
