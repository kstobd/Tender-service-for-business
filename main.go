package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Tender struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ServiceType     string    `json:"serviceType"`
	Status          string    `json:"status"`
	OrganizationID  int       `json:"organizationId"`
	CreatorUsername string    `json:"creatorUsername"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Bid struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	TenderID        uuid.UUID `json:"tenderId"`
	OrganizationID  int       `json:"organizationId"`
	CreatorUsername string    `json:"creatorUsername"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

var tenders = make(map[uuid.UUID]Tender)
var bids = make(map[uuid.UUID]Bid)

func main() {
	router := mux.NewRouter()

	// Эндпоинты для тендеров
	router.HandleFunc("/api/ping", PingHandler).Methods("GET")
	router.HandleFunc("/api/tenders", GetTenders).Methods("GET")
	router.HandleFunc("/api/tenders/new", CreateTender).Methods("POST")
	router.HandleFunc("/api/tenders/my", GetUserTenders).Methods("GET")
	router.HandleFunc("/api/tenders/{tenderId}/edit", EditTender).Methods("PATCH")
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", RollbackTender).Methods("PUT")

	// Эндпоинты для предложений
	router.HandleFunc("/api/bids/new", CreateBid).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func GetTenders(w http.ResponseWriter, r *http.Request) {
	var tenderList []Tender

	serviceType := r.URL.Query().Get("serviceType")
	for _, tender := range tenders {
		if serviceType == "" || tender.ServiceType == serviceType {
			tenderList = append(tenderList, tender)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenderList)
}

func CreateTender(w http.ResponseWriter, r *http.Request) {
	var tender Tender
	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	tender.ID = uuid.New()
	tender.Status = "Open"
	tender.CreatedAt = time.Now()
	tender.UpdatedAt = time.Now()

	tenders[tender.ID] = tender

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender)
}

func GetUserTenders(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	var userTenders []Tender

	for _, tender := range tenders {
		if tender.CreatorUsername == username {
			userTenders = append(userTenders, tender)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userTenders)
}

func EditTender(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		http.Error(w, "Invalid tender ID", http.StatusBadRequest)
		return
	}

	var updatedTender Tender
	err = json.NewDecoder(r.Body).Decode(&updatedTender)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	tender, exists := tenders[tenderID]
	if !exists {
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	tender.Name = updatedTender.Name
	tender.Description = updatedTender.Description
	tender.UpdatedAt = time.Now()

	tenders[tenderID] = tender

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender)
}

func RollbackTender(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		http.Error(w, "Invalid tender ID", http.StatusBadRequest)
		return
	}

	version := vars["version"]
	tender, exists := tenders[tenderID]
	if !exists {
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	tender.Name = tender.Name + " версия " + version

	tenders[tenderID] = tender

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender)
}

func CreateBid(w http.ResponseWriter, r *http.Request) {
	var bid Bid
	err := json.NewDecoder(r.Body).Decode(&bid)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Генерация ID и сохранение предложения
	bid.ID = uuid.New()
	bid.CreatedAt = time.Now()
	bid.UpdatedAt = time.Now()

	bids[bid.ID] = bid

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid)
}
