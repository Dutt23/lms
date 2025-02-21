package model

import "time"

type Member struct {
	Audited
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	JoinDate time.Time `json:"join_date"`
}
