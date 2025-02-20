package service

import (
	"context"
	"time"

	"github.com/dutt23/lms/model"
	"golang.org/x/exp/rand"
	"gopkg.in/redis.v5"
)


type BookCacheService interface {
  AddBook(c context.Context, book *model.Book) error
}

type bookCacheService struct {
  client *redis.Client
}

func NewBookCacheService(client *redis.Client) BookCacheService {
  rand.Seed(uint64(time.Now().UnixNano()))
  return &bookCacheService { client }
}

func (cache *bookCacheService) AddBook(c context.Context, book *model.Book) error {
  bookKey := CacheKey(c, "SET", string(book.Id))
  bookCountKey := CacheKey(c, "SET", book.Isbn)

  pipe := cache.client.Pipeline()
  bookExpiryTime := 1 * time.Hour
  jitter := time.Duration(rand.Int63n(int64(bookExpiryTime)))
  pipe.Set(bookKey, book, bookExpiryTime + jitter/2)

  bookCountExpiryTime := 4 * time.Hour
  jitter = time.Duration(rand.Int63n(int64(bookCountExpiryTime)))
  pipe.Set(bookCountKey, book, bookCountExpiryTime + jitter/2)
  _, err := pipe.Exec()
  return err
}