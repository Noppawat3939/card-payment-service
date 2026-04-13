package database

import (
	"card-payment-service/internal/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetPostgreslDSN()
	db, e := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if e != nil {
		return nil, e
	}

	log.Println("connected to database")
	return db, nil
}

func CloseDB(db *gorm.DB) {
	sql, e := db.DB()
	if e != nil {
		log.Printf("failed to get sql db for closing: %v", e)
		return
	}

	if e := sql.Close(); e != nil {
		log.Printf("failed to close db: %v", e)
		return
	}
}
