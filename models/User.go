package models

import "time"

type User struct {
	Id        uint      `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Username  string    `json:"username" gorm:"unique"'`
	Email     string    `json:"email" gorm:"unique"`
	Activated bool      `json:"activated"`
	Password  []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login"`
}
