package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

var conn *pgx.Conn

func main() {
	// Загрузка переменных окружения
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Подключение к PostgreSQL
	postgresConn := os.Getenv("POSTGRES_CONN")
	conn, err = pgx.Connect(context.Background(), postgresConn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Настройка HTTP сервера
	serverAddress := os.Getenv("SERVER_ADDRESS")
	router := mux.NewRouter()
	router.HandleFunc("/api/ping", PingHandler).Methods("GET")
	router.HandleFunc("/api/tenders", GetTendersHandler).Methods("GET")
	router.HandleFunc("/api/tenders/new", CreateTenderHandler).Methods("POST")

	log.Printf("Server is running at %s", serverAddress)
	err = http.ListenAndServe(serverAddress, router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Получение списка тендеров
func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := conn.Query(context.Background(), "SELECT id, name, description FROM tenders")
	if err != nil {
		http.Error(w, "Failed to retrieve tenders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tenders []map[string]interface{}

	for rows.Next() {
		var id, name, description string
		err = rows.Scan(&id, &name, &description)
		if err != nil {
			http.Error(w, "Failed to scan tenders", http.StatusInternalServerError)
			return
		}
		tenders = append(tenders, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)
}

// Создание нового тендера
func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	var tender struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		ServiceType     string `json:"serviceType"`
		OrganizationId  string `json:"organizationId"`
		CreatorUsername string `json:"creatorUsername"`
	}

	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO tenders (name, description, service_type, organization_id, creator_username)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id string
	err = conn.QueryRow(context.Background(), query, tender.Name, tender.Description, tender.ServiceType, tender.OrganizationId, tender.CreatorUsername).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to create tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":          id,
		"name":        tender.Name,
		"description": tender.Description,
	})
}
