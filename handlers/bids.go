package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"avito-project/db"
)

// CreateBidHandler обрабатывает создание нового предложения
func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("CreateBidHandler: Creating a new bid")

	// Структура для данных предложения
	var bid struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TenderID    string `json:"tenderId"`
		AuthorType  string `json:"authorType"`
		AuthorID    string `json:"authorId"`
	}

	// Декодирование JSON тела запроса
	err := json.NewDecoder(r.Body).Decode(&bid)
	if err != nil {
		log.Printf("CreateBidHandler: Invalid input: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	conn := db.GetConnection()

	// Проверка существования пользователя
	var userExists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM employee WHERE id = $1)", bid.AuthorID).Scan(&userExists)
	if err != nil {
		log.Printf("CreateBidHandler: Failed to check user existence: %v", err)
		http.Error(w, "Failed to check user existence", http.StatusInternalServerError)
		return
	}
	if !userExists {
		log.Printf("CreateBidHandler: User not found with id: %s", bid.AuthorID)
		http.Error(w, "User not found", http.StatusUnauthorized) // 401 ошибка
		return
	}

	// Проверка существования тендера
	var tenderExists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM tender WHERE id=$1)", bid.TenderID).Scan(&tenderExists)
	if err != nil {
		log.Printf("CreateBidHandler: Failed to check tender existence: %v", err)
		http.Error(w, "Failed to check tender existence", http.StatusInternalServerError)
		return
	}
	if !tenderExists {
		log.Println("CreateBidHandler: Tender not found")
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	// Статус по умолчанию для нового предложения
	status := "CREATED"
	version := 1
	createdAt := time.Now()

	// Вставка нового предложения в базу данных
	query := `
		INSERT INTO bids (name, description, status, tender_id, author_type, author_id, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		RETURNING id, created_at
	`

	var bidID string
	err = conn.QueryRow(context.Background(), query,
		bid.Name, bid.Description, status, bid.TenderID, bid.AuthorType, bid.AuthorID, version).
		Scan(&bidID, &createdAt)
	if err != nil {
		log.Printf("CreateBidHandler: Failed to create bid: %v", err)
		http.Error(w, "Failed to create bid", http.StatusInternalServerError)
		return
	}

	// Формируем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          bidID,
		"name":        bid.Name,
		"description": bid.Description,
		"status":      status,
		"tenderId":    bid.TenderID,
		"authorType":  bid.AuthorType,
		"authorId":    bid.AuthorID,
		"version":     version,
		"createdAt":   createdAt.Format(time.RFC3339),
	})

	log.Printf("CreateBidHandler: Bid created successfully in %v", time.Since(start))
}
