package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"gorm.io/gorm/clause"
)

type loanService struct {
	db connectors.SqliteConnector
}

func NewLoanService(db connectors.SqliteConnector) LoanService {
	return &loanService{db}
}

func (service *loanService) SaveLoan(ctx context.Context, memberId, bookId uint64, returnDate *time.Time) (*model.Loan, error) {
	t := time.Now()

	loan := &model.Loan{
		BookId:   bookId,
		MemberId: memberId,
		LoanDate: t,
	}

	//TODO: Add retry logic here
	if err := service.db.DB(ctx).Create(loan).Error; err != nil {
		return nil, err
	}
	return loan, nil
}

func (service *loanService) GetLoan(ctx context.Context, loanId uint64) (*model.Loan, error) {
	db := service.db.DB(ctx)
	var loan *model.Loan
	if err := db.Last(&loan, loanId).Error; err != nil {
		return nil, err
	}

	return loan, nil
}

func (service *loanService) GetLoans(ctx context.Context, lastId uint64, pageSize int) ([]*model.Loan, error) {
	db := service.db.DB(ctx)
	var loans []*model.Loan
	qry := db.Model(model.Loan{}).Where("id > ?", lastId).Limit(pageSize)

	tx := qry.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "published_date"},
		Desc:   true,
	}).Find(&loans)

	if tx.Error != nil {
		fmt.Println("not able to find any loans", tx.Error)
		return nil, tx.Error
	}
	return loans, nil
}

func (service *loanService) CompleteLoan(ctx context.Context, loanId uint64) error {
	db := service.db.DB(ctx)
	var loan *model.Loan
	if err := db.Last(&loan, loanId).Error; err != nil {
		return err
	}

	loan.ReturnDate = time.Now()
	return db.Save(loan).Error
}

func (service *loanService) DeleteLoan(ctx context.Context, loanId uint64) error {
	db := service.db.DB(ctx)
	return db.Delete(&model.Loan{}, loanId).Error
}
