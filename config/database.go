package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Message struct {
	gorm.Model
	Price     float64
	Denom     string
	Timestamp time.Time
	NodeID    string // Track the origin of the message
}

//var db *gorm.DB

//func InitDb() error {
//	var err error
//	db, err = gorm.Open(sqlite.Open("libp2p.db"), &gorm.Config{})
//	if err != nil {
//		return err
//	}
//	return db.AutoMigrate(&Message{})
//}

func initDb(postgresDsn string) (*gorm.DB, error) {
	// use postgres
	var database *gorm.DB
	var err error

	if postgresDsn[:8] == "postgres" {
		database, err = gorm.Open(postgres.Open(postgresDsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		database, err = gorm.Open(sqlite.Open(postgresDsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	database.AutoMigrate(&Message{})

	if err != nil {
		return nil, err
	}
	return database, nil
}
