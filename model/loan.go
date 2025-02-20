package model

type Loan struct {
	Audited
	BookId     int64       `json:"book_id"`
	MemberId   int64       `json:"member_id"`
	LoadDate   TimeWrapper `json:"loan_date"`
	ReturnDate TimeWrapper `json:"return_date"`
}
