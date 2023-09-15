package storage

import (
	"log"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	db  *gorm.DB
	cfg config.AppConfig
}

func InitDB(cfg config.AppConfig) *PostgresDB {

	if cfg.DataBaseString == "" {
		return nil
	}

	db, err := gorm.Open(postgres.Open(cfg.DataBaseString))
	if err != nil {
		log.Println(err)
		return nil
	}

	db.AutoMigrate(models.User{})

	return &PostgresDB{db: db,
		cfg: cfg}
}
