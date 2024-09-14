package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"avito-project/db"

	"github.com/gorilla/mux"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Ping endpoint hit")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("GetTendersHandler: Retrieving list of tenders")

	conn := db.GetConnection()
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

	conn := db.GetConnection()

	var creatorID string
	err = conn.QueryRow(context.Background(), "SELECT id FROM employee WHERE username = $1", tender.CreatorUsername).Scan(&creatorID)
	if err != nil {
		log.Printf("CreateTenderHandler: User not found: %v", err)
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	var responsibleID string
	err = conn.QueryRow(context.Background(), `
		SELECT id FROM organization_responsible
		WHERE organization_id = $1 AND user_id = $2`, tender.OrganizationId, creatorID).Scan(&responsibleID)
	if err != nil {
		log.Printf("CreateTenderHandler: User is not responsible for this organization: %v", err)
		http.Error(w, "User is not responsible for this organization", http.StatusForbidden)
		return
	}

	query := `INSERT INTO tender (name, description, service_type, organization_id, creator_id, status)
			  VALUES ($1, $2, $3, $4, $5, 'CREATED') RETURNING id`
	var tenderID string
	err = conn.QueryRow(context.Background(), query, tender.Name, tender.Description, tender.ServiceType, tender.OrganizationId, creatorID).Scan(&tenderID)
	if err != nil {
		log.Printf("CreateTenderHandler: Failed to create tender: %v", err)
		http.Error(w, "Failed to create tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":          tenderID,
		"name":        tender.Name,
		"description": tender.Description,
		"serviceType": tender.ServiceType,
		"status":      "CREATED",
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

	conn := db.GetConnection()

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
