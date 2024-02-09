package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=UTC",
		os.Getenv("YAMB_PQHOST"),
		os.Getenv("YAMB_PQUSERNAME"),
		os.Getenv("YAMB_PQPASSWORD"),
		"yamb",
	)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	log.Println("connected to database")
}
