package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

// GetUserBidsHandler обрабатывает получение списка предложений текущего пользователя с пагинацией
func GetUserBidsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Получение параметра username из query
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("GetUserBidsHandler: Username is required")
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Получение параметров пагинации: limit и offset
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Установка значений по умолчанию для limit и offset
	limit := 10
	offset := 0

	// Если переданы параметры пагинации, пытаемся их конвертировать
	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	log.Printf("GetUserBidsHandler: Retrieving bids for user %s with limit %d and offset %d", username, limit, offset)

	conn := db.GetConnection()

	// Проверка, существует ли пользователь
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("GetUserBidsHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized) // 401 ошибка
		return
	}

	// Запрос для получения списка предложений с использованием пагинации
	rows, err := conn.Query(context.Background(), `
		SELECT id, name, description, status, tender_id, author_type, version, created_at
		FROM bids
		WHERE author_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		log.Printf("GetUserBidsHandler: Failed to retrieve bids: %v", err)
		http.Error(w, "Failed to retrieve bids", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bids []map[string]interface{}

	// Обрабатываем строки результата
	for rows.Next() {
		var id, name, description, status, tenderID, authorType string
		var version int
		var createdAt time.Time
		err = rows.Scan(&id, &name, &description, &status, &tenderID, &authorType, &version, &createdAt)
		if err != nil {
			log.Printf("GetUserBidsHandler: Failed to scan bid: %v", err)
			http.Error(w, "Failed to scan bids", http.StatusInternalServerError)
			return
		}
		bids = append(bids, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"status":      status,
			"tenderId":    tenderID,
			"authorType":  authorType,
			"version":     version,
			"createdAt":   createdAt.Format(time.RFC3339),
		})
	}

	// Если не найдено предложений
	if len(bids) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	// Возвращаем список предложений
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)

	log.Printf("GetUserBidsHandler: Successfully retrieved bids for user %s in %v", username, time.Since(start))
}

// GetBidsForTenderHandler обрабатывает получение списка предложений для конкретного тендера
func GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Получение параметра tenderId из URL path
	tenderID := r.URL.Path[len("/api/bids/") : len(r.URL.Path)-len("/list")]
	if tenderID == "" {
		log.Println("GetBidsForTenderHandler: Tender ID is required")
		http.Error(w, "Tender ID is required", http.StatusBadRequest)
		return
	}

	// Получение параметра username из query
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("GetBidsForTenderHandler: Username is required")
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Получение параметров пагинации: limit и offset
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Установка значений по умолчанию для limit и offset
	limit := 10
	offset := 0

	// Если переданы параметры пагинации, пытаемся их конвертировать
	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	log.Printf("GetBidsForTenderHandler: Retrieving bids for tender %s and user %s with limit %d and offset %d", tenderID, username, limit, offset)

	conn := db.GetConnection()

	// Проверка, существует ли пользователь
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("GetBidsForTenderHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка, существует ли тендер
	var existingTenderID string
	err = conn.QueryRow(context.Background(), "SELECT id FROM tender WHERE id = $1", tenderID).Scan(&existingTenderID)
	if err != nil {
		log.Printf("GetBidsForTenderHandler: Tender not found: %v", err)
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	// Проверка прав пользователя (например, является ли он ответственным за этот тендер)
	var isResponsible bool
	err = conn.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 
			FROM organization_responsible AS orp
			INNER JOIN tender ON tender.organization_id = orp.organization_id
			WHERE orp.user_id = $1 AND tender.id = $2
		)`, userID, tenderID).Scan(&isResponsible)
	if err != nil {
		log.Printf("GetBidsForTenderHandler: Error checking user permissions: %v", err)
		http.Error(w, "Error checking user permissions", http.StatusInternalServerError)
		return
	}

	if !isResponsible {
		log.Printf("GetBidsForTenderHandler: User %s does not have permission for tender %s", username, tenderID)
		http.Error(w, "User does not have permission for this tender", http.StatusForbidden)
		return
	}

	// Запрос списка предложений для тендера с учетом пагинации
	rows, err := conn.Query(context.Background(), `
		SELECT id, name, description, status, author_type, version, created_at
		FROM bids
		WHERE tender_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`, tenderID, limit, offset)
	if err != nil {
		log.Printf("GetBidsForTenderHandler: Failed to retrieve bids: %v", err)
		http.Error(w, "Failed to retrieve bids", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bids []map[string]interface{}

	// Обработка результатов запроса
	for rows.Next() {
		var id, name, description, status, authorType string
		var version int
		var createdAt time.Time
		err = rows.Scan(&id, &name, &description, &status, &authorType, &version, &createdAt)
		if err != nil {
			log.Printf("GetBidsForTenderHandler: Failed to scan bid: %v", err)
			http.Error(w, "Failed to scan bids", http.StatusInternalServerError)
			return
		}
		bids = append(bids, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"status":      status,
			"authorType":  authorType,
			"version":     version,
			"createdAt":   createdAt.Format(time.RFC3339),
		})
	}

	// Если не найдено предложений
	if len(bids) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	// Возвращаем список предложений
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)

	log.Printf("GetBidsForTenderHandler: Successfully retrieved bids for tender %s in %v", tenderID, time.Since(start))
}

func SubmitBidDecisionHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("SubmitBidDecisionHandler: Processing decision submission")

	// Получение параметра tenderId из URL path
	bidID := r.URL.Path[len("/api/bids/") : len(r.URL.Path)-len("/submit_decision")]
	if bidID == "" {
		log.Println("SubmitBidDecisionHandler: Bid ID is required")
		http.Error(w, "Bid ID is required", http.StatusBadRequest)
		return
	}

	// Получение параметров из запроса
	decision := r.URL.Query().Get("decision")
	username := r.URL.Query().Get("username")

	if decision == "" || username == "" {
		log.Println("SubmitBidDecisionHandler: Missing required parameters")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Валидация решения
	if decision != "Approved" && decision != "Rejected" {
		log.Printf("SubmitBidDecisionHandler: Invalid decision: %s", decision)
		http.Error(w, "Invalid decision value", http.StatusBadRequest)
		return
	}

	conn := db.GetConnection()

	// Проверка существования пользователя
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("SubmitBidDecisionHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка существования предложения
	var currentBidStatus string
	err = conn.QueryRow(context.Background(), "SELECT status FROM bids WHERE id = $1", bidID).Scan(&currentBidStatus)
	if err != nil {
		log.Printf("SubmitBidDecisionHandler: Bid not found: %v", err)
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	// Проверка прав доступа
	var isAuthorized bool
	err = conn.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM tender 
			WHERE creator_id = $1 
			AND id = (SELECT tender_id FROM bids WHERE id = $2)
		)`, userID, bidID).Scan(&isAuthorized)
	if err != nil || !isAuthorized {
		log.Printf("SubmitBidDecisionHandler: User is not authorized to submit decision for this bid: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Получаем tenderID
	var tenderID string
	err = conn.QueryRow(context.Background(), `
		SELECT tender_id FROM bids
		WHERE id = $1`, bidID).Scan(&tenderID)
	if err != nil {
		log.Printf("SubmitBidDecisionHandler: Tender not found: %v", err)
		http.Error(w, "Tender not found", http.StatusForbidden)
		return
	}

	// Если решение "Approved", закрываем тендер
	if decision == "Approved" {
		_, err = conn.Exec(context.Background(), `
			UPDATE tender
			SET status = 'CLOSED'
			WHERE id = $1`, tenderID)
		if err != nil {
			log.Printf("SubmitBidDecisionHandler: Failed to close tender: %v", err)
			http.Error(w, "Failed to close tender", http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем обновленные данные предложения
	var bid struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		Version     int       `json:"version"`
		CreatedAt   time.Time `json:"createdAt"`
	}

	err = conn.QueryRow(context.Background(), `
		SELECT id, name, description, status, version, created_at
		FROM bids
		WHERE id = $1`, bidID).Scan(&bid.ID, &bid.Name, &bid.Description, &bid.Status, &bid.Version, &bid.CreatedAt)
	if err != nil {
		log.Printf("SubmitBidDecisionHandler: Failed to retrieve updated bid: %v", err)
		http.Error(w, "Failed to retrieve updated bid", http.StatusInternalServerError)
		return
	}

	// Ответ с данными обновленного предложения
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid)

	log.Printf("SubmitBidDecisionHandler: Decision submitted successfully for bid %s in %v", bidID, time.Since(start))
}
