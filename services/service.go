package service

import (
	"context"

	"github.com/dutt23/lms/model"
)


type BookService interface {
  GetBook(ctx context.Context, bookId uint64) (*model.Book, error)
}
type MemberService interface {
   GetMember(ctx context.Context, memberId uint64) (*model.Member, error)
}