package gormClient

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

// Maximum number of connection
const maxAttempts = 10

// NewClient creates a new MySQL client using GORM.
// It keeps attempting to connect to the database until successful, or until maxAttempts has been reached.
// In case of a failure, it returns an error.
func NewClient() (*gorm.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Attempt %d: MySQL not ready, backing off for two seconds...", attempt)
			time.Sleep(2 * time.Second)
		} else {
			log.Println("Connected to database!")
			return db, nil
		}
	}

	return nil, fmt.Errorf("connection attempts exceeded: could not connect to MySQL")
}
