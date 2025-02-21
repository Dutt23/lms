package model

import "time"

type Book struct {
	Audited
	Title           string    `json:"title" gorm:"type:string"`
	Author          string    `json:"author" gorm:"type:string"`
	PublishedDate   time.Time `json:"published_date"`
	Isbn            string    `json:"isbn"`
	NumberOfPages   uint64    `json:"number_of_pages"`
	CoverImage      string    `json:"cover_image"`
	Language        string    `json:"language"`
	AvailableCopies int64     `json:"available_copies"`
}
