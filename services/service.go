package service

import (
	"context"
	"time"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/model"
)

type BookService interface {
	GetBook(ctx context.Context, bookId uint64) (*model.Book, error)
	ChangeAvailableCopies(ctx context.Context, bookId uint64, count int64) error
	GetBooks(ctx context.Context, lastId uint64, pageSize int) ([]*model.Book, error)
}
type MemberService interface {
	GetMember(ctx context.Context, memberId uint64) (*model.Member, error)
}

type LoanService interface {
	SaveLoan(ctx context.Context, memberId, bookId uint64, returnDate *time.Time) (*model.BookLoan, error)
	GetLoan(ctx context.Context, loanId uint64) (*model.BookLoan, error)
	GetLoans(ctx context.Context, lastId uint64, pageSize int) ([]*model.BookLoan, error)
	CompleteLoan(ctx context.Context, loanId uint64) error
	DeleteLoan(ctx context.Context, loanId uint64) error
}

type AnalyticsService interface {
	GetBookListAnalytics(ctx context.Context, bookIds []uint64) (*cache.BookAnalytics, error)
}
