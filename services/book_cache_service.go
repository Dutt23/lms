package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dutt23/lms/model"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/rand"
)

const (
  BOOK_ID_FILTER = "books:id"
)
type BookCacheService interface {
  StoreBookMetaInCache(c context.Context, book *model.Book) error
}

type bookCacheService struct {
  client *redis.Client
}


func NewBookCacheService(client *redis.Client) BookCacheService {
  return &bookCacheService { client }
}

func (cache *bookCacheService) StoreBookMetaInCache(c context.Context, book *model.Book) error {
  bookKey := CacheKey(c, "SET", string(book.Id))
  bookCountKey := CacheKey(c, "SET", book.Isbn)

  pipe := cache.client.Pipeline()
  bookExpiryTime := 1 * time.Hour
  jitter := time.Duration(rand.Int63n(int64(bookExpiryTime)))
  pipe.Set(c, bookKey, book, bookExpiryTime + jitter/2)

  bookCountExpiryTime := 4 * time.Hour
  jitter = time.Duration(rand.Int63n(int64(bookCountExpiryTime)))
  pipe.Set(c, bookCountKey, book, bookCountExpiryTime + jitter/2)

  pipe.BFAdd(c, BOOK_ID_FILTER, book.Id)

  _, err := pipe.Exec(c)
  return err
}

func (cache *bookCacheService) DoesBookExist(c context.Context, bookId uint64) bool {
  res, err := cache.client.BFExists(c, BOOK_ID_FILTER, bookId).Result()

  if err != nil {
    fmt.Errorf("unable to determine result %w", err)
    // This will go to the database for confirmation
    return true
  }
  return res
}