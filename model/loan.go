package model

type Loan struct {
	Audited
	BookId     uint64      `json:"book_id"`
	MemberId   uint64      `json:"member_id"`
	LoanDate   TimeWrapper `json:"loan_date"`
	ReturnDate TimeWrapper `json:"return_date"`
}
