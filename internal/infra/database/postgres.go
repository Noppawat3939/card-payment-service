package database

import (
	"card-payment-service/internal/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetPostgreslDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	log.Println("connected to database")
	return db, nil
}

func CloseDB(db *gorm.DB) {
	sql, err := db.DB()
	if err != nil {
		log.Printf("failed to get sql db for closing: %v", err)
		return
	}

	if err := sql.Close(); err != nil {
		log.Printf("failed to close db: %v", err)
		return
	}
}
