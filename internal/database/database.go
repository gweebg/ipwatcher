package database

import (
	"github.com/gweebg/ipwatcher/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func ConnectDatabase() {
	var err error

	db, err = gorm.Open(sqlite.Open("watcher.db"), &gorm.Config{})
	utils.Check(err, "")
}

func GetDatabase() *gorm.DB {
	return db
}
