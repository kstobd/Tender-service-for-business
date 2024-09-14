package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv загружает переменные окружения из файла .env
func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Environment loaded successfully")
	return nil
}
