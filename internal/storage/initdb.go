package storage

import (
	"log"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg config.AppConfig) *gorm.DB {

	db, err := gorm.Open(postgres.Open(cfg.DataBaseString), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database!")
	}

	if cfg.DataBaseString == "" {
		return nil
	}

	users := db.AutoMigrate(models.User{})
	if users != nil {
		log.Fatal(err)
		return nil
	}

	orders := db.AutoMigrate(models.Orders{})
	if orders != nil {
		log.Fatal(err)
		return nil
	}

	return db
}
