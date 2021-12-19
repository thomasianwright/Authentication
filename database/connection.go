package database

import (
	"authapi/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	conn, err := gorm.Open(mysql.Open("root:rootroot@/auth"), &gorm.Config{})

	DB = conn

	if err != nil {
		panic("Could not connect to the database")
	}

	conn.AutoMigrate(&models.User{})
}
