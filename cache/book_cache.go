package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"golang.org/x/exp/rand"
)

type BookCache interface {
	StoreBookMetaInCache(c context.Context, book *model.Book) error
	DoesBookExist(c context.Context, bookId uint64) bool
	GetBook(c context.Context, bookId uint64) (*model.Book, error)
	DeleteBook(c context.Context, bookId uint64) error
	IsIsbnUnique(c context.Context, isbn string) bool
}

type bookCache struct {
	conn connectors.CacheConnector
}

func NewBookCache(client connectors.CacheConnector) BookCache {
	return &bookCache{conn: client}
}

func (cache *bookCache) StoreBookMetaInCache(c context.Context, book *model.Book) error {
	bookKey := CacheKey(c, "SET_BOOK", fmt.Sprintf("%d", book.Id))
	bookCountKey := CacheKey(c, "SET_BOOK", book.Isbn)

	db := cache.conn.DB(c)
	pipe := db.Pipeline()
	bookExpiryTime := 1 * time.Hour
	jitter := time.Duration(rand.Int63n(int64(bookExpiryTime)))
	data, err := json.Marshal(book)
	if err != nil {
		fmt.Errorf("Unable to cache the record as value is not marshalable %s", err, bookKey)
	}
	pipe.Set(c, bookKey, data, bookExpiryTime+jitter/2)

	bookCountExpiryTime := 4 * time.Hour
	jitter = time.Duration(rand.Int63n(int64(bookCountExpiryTime)))
	pipe.Set(c, bookCountKey, book.AvailableCopies, bookCountExpiryTime+jitter/2)

	pipe.BFAdd(c, BOOK_ISBN_FILTER, book.Id)

	_, err = pipe.Exec(c)
	return err
}

func (cache *bookCache) DoesBookExist(c context.Context, bookId uint64) bool {
	db := cache.conn.DB(c)
	bookKey := CacheKey(c, "SET_BOOK", fmt.Sprintf("%d", bookId))
	res, err := db.Exists(c, bookKey).Result()

	if err != nil {
		fmt.Errorf("unable to determine result %w", err)
		// This will go to the database for confirmation
		return true
	}
	return res > 0
}

func (cache *bookCache) IsIsbnUnique(c context.Context, isbn string) bool {
	db := cache.conn.DB(c)
	res, err := db.BFExists(c, BOOK_ISBN_FILTER, isbn).Result()

	if err != nil {
		fmt.Errorf("unable to determine result %w", err)
		// This will go to the database for confirmation
		return true
	}
	return !res
}

func (cache *bookCache) DeleteBook(c context.Context, bookId uint64) error {
	db := cache.conn.DB(c)
	bookKey := CacheKey(c, "SET_BOOK", fmt.Sprintf("%d", bookId))
	return db.Del(c, bookKey).Err()
}

func (cache *bookCache) GetBook(c context.Context, bookId uint64) (*model.Book, error) {
	db := cache.conn.DB(c)
	bookKey := CacheKey(c, "SET_BOOK", fmt.Sprintf("%d", bookId))
	res, err := db.Get(c, bookKey).Bytes()

	if err != nil {
		fmt.Println(fmt.Errorf("unable to get result from cache %w", err))
		// This will go to the database for confirmation
		return nil, err
	}

	book := &model.Book{}
	err = json.Unmarshal(res, book)

	if err != nil {
		fmt.Println(fmt.Errorf("unable to get result from cache %w", err))
		// This will go to the database for confirmation
		return nil, err
	}
	return book, nil
}
