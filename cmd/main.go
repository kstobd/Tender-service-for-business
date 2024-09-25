package main

import (
	"log"
	"net/http"
	"os"

	"avito-project/db"
	"avito-project/routes"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting server setup...")

	// УБРАТЬ КОММЕНТИРОВАНИЕ, ЕСЛИ ИСПОЛЬЗВУЕТСЯ ФАЙЛ .env!!!!

	// // Загрузка переменных окружения
	// err := config.LoadEnv()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	// Подключение к PostgreSQL
	db.Connect()
	defer db.Close()

	// Запуск миграций
	db.RunMigrations()

	// Настройка маршрутизации
	serverAddress := os.Getenv("SERVER_ADDRESS")
	router := mux.NewRouter()
	routes.SetupRoutes(router)

	log.Printf("Server is running at %s", serverAddress)
	err := http.ListenAndServe(serverAddress, router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
