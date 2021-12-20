package database

import (
	"authapi/models"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func Connect() {
	conn, err := gorm.Open(mysql.Open(os.Getenv("connection")+"?parseTime=true"), &gorm.Config{})

	DB = conn

	if err != nil {
		panic("Could not connect to the database")
	}

	err = conn.AutoMigrate(&models.User{}, &models.Activation{})
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
}
