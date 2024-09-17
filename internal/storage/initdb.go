package storage

import (
	"log"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg config.AppConfig) *gorm.DB {

	db, err := gorm.Open(postgres.Open(cfg.DataBaseString), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})

	if err != nil {
		log.Fatal("Failed to connect to database!")
	}

	if cfg.DataBaseString == "" {
		return nil
	}

	users := db.AutoMigrate(models.Users{})
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
