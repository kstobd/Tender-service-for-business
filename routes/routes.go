package routes

import (
	"avito-project/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/api/ping", handlers.PingHandler).Methods("GET")
	router.HandleFunc("/api/tenders", handlers.GetTendersHandler).Methods("GET")
	router.HandleFunc("/api/tenders/new", handlers.CreateTenderHandler).Methods("POST")
	router.HandleFunc("/api/tenders/my", handlers.GetMyTendersHandler).Methods("GET")
	router.HandleFunc("/api/tenders/{tenderId}/status", handlers.GetTenderStatusHandler).Methods("GET")
	router.HandleFunc("/api/tenders/{tenderId}/status", handlers.UpdateTenderStatusHandler).Methods("PUT")
	router.HandleFunc("/api/tenders/{tenderId}/edit", handlers.EditTenderHandler).Methods("PATCH")
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", handlers.RollbackTenderHandler).Methods("PUT")
}
