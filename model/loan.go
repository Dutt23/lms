package model

import "time"

type BookLoan struct {
	Audited
	BookId     uint64    `json:"book_id"`
	MemberId   uint64    `json:"member_id"`
	LoanDate   time.Time `json:"loan_date"`
	ReturnDate time.Time `json:"return_date"`
}
