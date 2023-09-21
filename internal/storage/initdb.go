package storage

import (
	"log"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg config.AppConfig) *gorm.DB {

	if cfg.DataBaseString == "" {
		return nil
	}

	db, err := gorm.Open(postgres.Open(cfg.DataBaseString))
	if err != nil {
		log.Println(err)
		return nil
	}

	db.AutoMigrate(models.Users{})

	return db
}
