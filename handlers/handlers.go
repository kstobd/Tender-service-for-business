package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"avito-project/db"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Ping endpoint hit")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Tender represents a tender record
type Tender struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("GetTendersHandler: Retrieving list of tenders")

	// Получение фильтра из query-параметров
	serviceTypeFilter := r.URL.Query().Get("service_type")

	conn := db.GetConnection()
	var rows pgx.Rows
	var err error

	// Проверяем, передан ли фильтр по service_type
	if serviceTypeFilter != "" {
		// Если фильтр передан, добавляем условие WHERE в запрос
		rows, err = conn.Query(context.Background(),
			"SELECT id, name, description, service_type, status FROM tender WHERE service_type = $1", serviceTypeFilter)
	} else {
		// Если фильтр не передан, возвращаем все записи
		rows, err = conn.Query(context.Background(),
			"SELECT id, name, description, service_type, status FROM tender")
	}

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

	// Декодирование JSON тела запроса
	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		log.Printf("CreateTenderHandler: Invalid input: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	conn := db.GetConnection()

	// Проверка существования пользователя
	var creatorID string
	err = conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", tender.CreatorUsername).Scan(&creatorID)
	if err != nil {
		log.Printf("CreateTenderHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка, является ли пользователь ответственным за организацию
	var responsibleID string
	err = conn.QueryRow(context.Background(), `
		SELECT id FROM organization_responsible
		WHERE organization_id = $1 AND user_id = $2`, tender.OrganizationId, creatorID).Scan(&responsibleID)
	if err != nil {
		log.Printf("CreateTenderHandler: User is not responsible for this organization: %v", err)
		http.Error(w, "User is not responsible for this organization", http.StatusForbidden)
		return
	}

	// Создание тендера
	query := `INSERT INTO tender (name, description, service_type, organization_id, creator_id, responsible_id, status, version, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, 'CREATED', 1, CURRENT_TIMESTAMP) RETURNING id, created_at`
	var tenderID string
	var createdAt time.Time
	err = conn.QueryRow(context.Background(), query,
		tender.Name, tender.Description, tender.ServiceType, tender.OrganizationId, creatorID, responsibleID).
		Scan(&tenderID, &createdAt)
	if err != nil {
		log.Printf("CreateTenderHandler: Failed to create tender: %v", err)
		http.Error(w, "Failed to create tender", http.StatusInternalServerError)
		return
	}

	// Ответ с данными созданного тендера
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          tenderID,
		"name":        tender.Name,
		"description": tender.Description,
		"serviceType": tender.ServiceType,
		"status":      "CREATED",
		"version":     1, // Начальная версия тендера
		"createdAt":   createdAt.Format(time.RFC3339),
	})

	log.Printf("CreateTenderHandler: Tender created successfully in %v", time.Since(start))
}

func GetMyTendersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("GetMyTendersHandler: Username is required")
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	log.Printf("GetMyTendersHandler: Retrieving tenders for user %s", username)

	conn := db.GetConnection()

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

func GetTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]
	username := r.URL.Query().Get("username")

	log.Printf("GetTenderStatusHandler: Getting status for tender %s", tenderId)

	conn := db.GetConnection()

	// Проверка существования пользователя
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("GetTenderStatusHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Получение статуса тендера
	var status string
	err = conn.QueryRow(context.Background(), "SELECT status FROM tender WHERE id = $1", tenderId).Scan(&status)
	if err != nil {
		log.Printf("GetTenderStatusHandler: Tender not found: %v", err)
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
	})

	log.Printf("GetTenderStatusHandler: Successfully retrieved status in %v", time.Since(start))
}

// UpdateTenderStatusHandler: Изменить статус тендера по ID
func UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]
	status := strings.ToUpper(r.URL.Query().Get("status"))
	username := r.URL.Query().Get("username")

	log.Printf("UpdateTenderStatusHandler: Updating status for tender %s", tenderId)

	conn := db.GetConnection()

	// Проверка существования пользователя
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("UpdateTenderStatusHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка прав пользователя
	var responsibleID string
	err = conn.QueryRow(context.Background(), `
		SELECT id FROM organization_responsible
		WHERE organization_id = (SELECT organization_id FROM tender WHERE id = $1)
		AND user_id = $2`, tenderId, userID).Scan(&responsibleID)
	if err != nil {
		log.Printf("UpdateTenderStatusHandler: User is not responsible for this tender: %v", err)
		http.Error(w, "User is not responsible for this tender", http.StatusForbidden)
		return
	}

	// Проверка допустимого статуса
	if status != "CREATED" && status != "PUBLISHED" && status != "CLOSED" {
		log.Printf("UpdateTenderStatusHandler: Invalid status value: %s", status)
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	// Обновление статуса тендера
	_, err = conn.Exec(context.Background(), "UPDATE tender SET status = $1 WHERE id = $2", status, tenderId)
	if err != nil {
		log.Printf("UpdateTenderStatusHandler: Failed to update status: %v", err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
	})

	log.Printf("UpdateTenderStatusHandler: Successfully updated status in %v", time.Since(start))
}

func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]
	username := r.URL.Query().Get("username")

	log.Printf("EditTenderHandler: Editing tender %s", tenderId)

	// Проверка существования пользователя
	conn := db.GetConnection()
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("EditTenderHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка прав пользователя на редактирование тендера
	var responsibleID string
	err = conn.QueryRow(context.Background(), `
		SELECT id FROM organization_responsible
		WHERE organization_id = (SELECT organization_id FROM tender WHERE id = $1)
		AND user_id = $2`, tenderId, userID).Scan(&responsibleID)
	if err != nil {
		log.Printf("EditTenderHandler: User is not responsible for this tender: %v", err)
		http.Error(w, "User is not responsible for this tender", http.StatusForbidden)
		return
	}

	// Декодирование запроса
	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Printf("EditTenderHandler: Invalid input: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Построение запроса на обновление
	var fields []string
	var values []interface{}
	idx := 1

	if name, ok := updates["name"].(string); ok {
		fields = append(fields, "name = $"+strconv.Itoa(idx))
		values = append(values, name)
		idx++
	}

	if description, ok := updates["description"].(string); ok {
		fields = append(fields, "description = $"+strconv.Itoa(idx))
		values = append(values, description)
		idx++
	}

	if serviceType, ok := updates["serviceType"].(string); ok {
		fields = append(fields, "service_type = $"+strconv.Itoa(idx))
		values = append(values, serviceType)
		idx++
	}

	if len(fields) == 0 {
		log.Println("EditTenderHandler: No fields to update")
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Инкремент версии и добавление в запрос
	fields = append(fields, "version = version + 1, updated_at = CURRENT_TIMESTAMP")
	query := "UPDATE tender SET " + strings.Join(fields, ", ") + " WHERE id = $" + strconv.Itoa(idx)
	values = append(values, tenderId)

	// Выполнение запроса
	_, err = conn.Exec(context.Background(), query, values...)
	if err != nil {
		log.Printf("EditTenderHandler: Failed to update tender: %v", err)
		http.Error(w, "Failed to update tender", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленную информацию о тендере
	var tender Tender
	err = conn.QueryRow(context.Background(), "SELECT id, name, description, service_type, status, version, created_at FROM tender WHERE id = $1", tenderId).Scan(
		&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.Version, &tender.CreatedAt,
	)
	if err != nil {
		log.Printf("EditTenderHandler: Failed to retrieve updated tender: %v", err)
		http.Error(w, "Failed to retrieve updated tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender)

	log.Printf("EditTenderHandler: Tender updated successfully in %v", time.Since(start))
}

// RollbackTenderHandler: Откат к предыдущей версии тендера
func RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]
	versionStr := vars["version"]
	username := r.URL.Query().Get("username")

	log.Printf("RollbackTenderHandler: Rolling back tender %s to version %s", tenderId, versionStr)

	// Проверка существования пользователя
	conn := db.GetConnection()
	var userID string
	err := conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", username).Scan(&userID)
	if err != nil {
		log.Printf("RollbackTenderHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка прав пользователя на откат тендера
	var responsibleID string
	err = conn.QueryRow(context.Background(), `
		SELECT id FROM organization_responsible
		WHERE organization_id = (SELECT organization_id FROM tender WHERE id = $1)
		AND user_id = $2`, tenderId, userID).Scan(&responsibleID)
	if err != nil {
		log.Printf("RollbackTenderHandler: User is not responsible for this tender: %v", err)
		http.Error(w, "User is not responsible for this tender", http.StatusForbidden)
		return
	}

	// Проверка существования версии
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		log.Printf("RollbackTenderHandler: Invalid version format: %v", err)
		http.Error(w, "Invalid version format", http.StatusBadRequest)
		return
	}

	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM tender_versions WHERE tender_id = $1 AND version = $2", tenderId, version).Scan(&count)
	if err != nil || count == 0 {
		log.Printf("RollbackTenderHandler: Version not found for tender %s and version %d", tenderId, version)
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	// Откат к указанной версии и инкремент версии
	_, err = conn.Exec(context.Background(), `
		UPDATE tender
		SET name = v.name, description = v.description, service_type = v.service_type, version = version + 1, updated_at = CURRENT_TIMESTAMP
		FROM tender_versions v
		WHERE tender.id = $1 AND v.tender_id = $1 AND v.version = $2`, tenderId, version)
	if err != nil {
		log.Printf("RollbackTenderHandler: Failed to rollback tender: %v", err)
		http.Error(w, "Failed to rollback tender", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленную информацию о тендере
	var tender Tender
	err = conn.QueryRow(context.Background(), "SELECT id, name, description, service_type, status, version, created_at FROM tender WHERE id = $1", tenderId).Scan(
		&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.Version, &tender.CreatedAt,
	)
	if err != nil {
		log.Printf("RollbackTenderHandler: Failed to retrieve updated tender: %v", err)
		http.Error(w, "Failed to retrieve updated tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender)

	log.Printf("RollbackTenderHandler: Tender rolled back successfully in %v", time.Since(start))
}
