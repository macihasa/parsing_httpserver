package routes

import (
	"github.com/gorilla/mux"
	"github.com/macihasa/parsing_httpserver/pkg/handlers"
)

func Init() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", handlers.Health).Methods("GET")
	r.HandleFunc("/dcecstmsmsg", handlers.DCCXML).Methods("GET", "POST")
	r.HandleFunc("/dcecstmsmsg/getfile", handlers.HandleCsvRequest).Methods("GET", "OPTIONS")
	r.HandleFunc("/excel/{sheet}", handlers.Excel).Methods("POST")

	return r
}
