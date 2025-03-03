package service

import (
	"context"
	"fmt"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"gorm.io/gorm/clause"
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

func (service *bookService) GetBooks(ctx context.Context, lastId uint64, pageSize int) ([]*model.Book, error) {
	db := service.db.DB(ctx)
	var books []*model.Book
	qry := db.Model(model.Book{}).Where("id > ?", lastId).Limit(pageSize)

	tx := qry.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "published_date"},
		Desc:   true,
	}).Find(&books)

	if tx.Error != nil {
		fmt.Println("not able to find any loans", tx.Error)
		return nil, tx.Error
	}
	return books, nil
}
