package models

type Activation struct {
	Id     uint   `json:"id"`
	UserId uint   `json:"userId"`
	Guid   string `json:"guid" gorm:"unique"`
}
