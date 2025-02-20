package model

type Book struct {
  Audited
  Title string `json:"title" gorm:"type:string"`
  Author string `json:"author" gorm:"type:string"`
  PublishedDate TimeWrapper `json:"published_date"`
  Isbn string `json:"isbn"`
  NumberOfPages int64 `json:"number_of_pages"`
  CoverImage string `json:"cover_image"`
  Language string `json:"language"`
  AvailableCopies int64 `json:"available_copies"`
}