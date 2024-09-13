package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

var conn *pgx.Conn

// проверяю, могу ли Я еще вносить изменения
func runMigrations() {
	log.Println("Starting database migrations...")

	postgresConn := os.Getenv("POSTGRES_CONN")
	db, err := sql.Open("postgres", postgresConn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations ran successfully")
}

func main() {
	log.Println("Starting server setup...")

	// Загрузка переменных окружения
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Подключение к PostgreSQL
	postgresConn := os.Getenv("POSTGRES_CONN")
	log.Printf("Connecting to database at %s", postgresConn)
	conn, err = pgx.Connect(context.Background(), postgresConn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())
	log.Println("Successfully connected to the database")

	// Запуск миграций
	runMigrations()

	// Настройка HTTP сервера
	serverAddress := os.Getenv("SERVER_ADDRESS")
	router := mux.NewRouter()
	router.HandleFunc("/api/ping", PingHandler).Methods("GET")
	router.HandleFunc("/api/tenders", GetTendersHandler).Methods("GET")
	router.HandleFunc("/api/tenders/new", CreateTenderHandler).Methods("POST")
	router.HandleFunc("/api/tenders/my", GetMyTendersHandler).Methods("GET")
	router.HandleFunc("/api/tenders/{tenderId}/edit", EditTenderHandler).Methods("PATCH")

	log.Printf("Server is running at %s", serverAddress)
	err = http.ListenAndServe(serverAddress, router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Ping endpoint hit")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Получение списка тендеров
func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("GetTendersHandler: Retrieving list of tenders")

	rows, err := conn.Query(context.Background(), "SELECT id, name, description, service_type, status FROM tender")
	if err != nil {
		log.Printf("GetTendersHandler: Failed to retrieve tenders: %v", err)
		http.Error(w, "Failed to retrieve tenders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tenders []map[string]interface{}

	for rows.Next() {
		var id, name, description, serviceType, status string
		err = rows.Scan(&id, &name, &description, &serviceType, &status)
		if err != nil {
			log.Printf("GetTendersHandler: Failed to scan tenders: %v", err)
			http.Error(w, "Failed to scan tenders", http.StatusInternalServerError)
			return
		}
		tenders = append(tenders, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"serviceType": serviceType,
			"status":      status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)

	log.Printf("GetTendersHandler: Successfully retrieved tenders in %v", time.Since(start))
}

// Создание нового тендера
func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("CreateTenderHandler: Creating a new tender")

	var tender struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		ServiceType     string `json:"serviceType"`
		OrganizationId  string `json:"organizationId"`
		CreatorUsername string `json:"creatorUsername"`
	}

	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		log.Printf("CreateTenderHandler: Invalid input: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var creatorID string
	err = conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", tender.CreatorUsername).Scan(&creatorID)
	if err != nil {
		log.Printf("CreateTenderHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO tender (name, description, service_type, organization_id, creator_id, status)
			  VALUES ($1, $2, $3, $4, $5, 'CREATED') RETURNING id`
	var id string
	err = conn.QueryRow(context.Background(), query, tender.Name, tender.Description, tender.ServiceType, tender.OrganizationId, creatorID).Scan(&id)
	if err != nil {
		log.Printf("CreateTenderHandler: Failed to create tender: %v", err)
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

	log.Printf("CreateTenderHandler: Tender created successfully in %v", time.Since(start))
}

// Получение тендеров текущего пользователя
func GetMyTendersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("GetMyTendersHandler: Username is required")
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	log.Printf("GetMyTendersHandler: Retrieving tenders for user %s", username)

	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("GetMyTendersHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	rows, err := conn.Query(context.Background(), "SELECT id, name, description FROM tender WHERE creator_id = $1", userID)
	if err != nil {
		log.Printf("GetMyTendersHandler: Failed to retrieve tenders: %v", err)
		http.Error(w, "Failed to retrieve tenders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tenders []map[string]interface{}

	for rows.Next() {
		var id, name, description string
		err = rows.Scan(&id, &name, &description)
		if err != nil {
			log.Printf("GetMyTendersHandler: Failed to scan tenders: %v", err)
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

	log.Printf("GetMyTendersHandler: Successfully retrieved tenders for user %s in %v", username, time.Since(start))
}

// Редактирование тендера
func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenderID := mux.Vars(r)["tenderId"]
	log.Printf("EditTenderHandler: Editing tender with ID %s", tenderID)

	var tender struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		log.Printf("EditTenderHandler: Invalid input: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `UPDATE tender SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err = conn.Exec(context.Background(), query, tender.Name, tender.Description, tenderID)
	if err != nil {
		log.Printf("EditTenderHandler: Failed to update tender: %v", err)
		http.Error(w, "Failed to update tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":          tenderID,
		"name":        tender.Name,
		"description": tender.Description,
	})

	log.Printf("EditTenderHandler: Tender with ID %s edited successfully in %v", tenderID, time.Since(start))
}
