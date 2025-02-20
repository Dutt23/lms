package model

type Member struct {
  Audited
  Name string `json:"name"`
  Email string `json:"email"`
  JoinDate TimeWrapper `json:"join_date"`
}