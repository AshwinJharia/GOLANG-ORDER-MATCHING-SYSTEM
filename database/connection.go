package database

import (
	"database/sql"
	"log"
	"order-matching-engine/config"
	"os"

	"github.com/joho/godotenv"
)

var DB *sql.DB

func InitDB() error {
	// Try to load .env file (optional - won't fail if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}
	// Get database configuration from environment variables with defaults
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password" // Default for development only
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		database = "order_matching"
	}

	dbConfig := config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}

	var err error
	DB, err = config.NewDatabaseConnection(dbConfig)
	return err
}