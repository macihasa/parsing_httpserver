package main

import (
	"log"
	"net/http"

	"github.com/macihasa/parsing_httpserver/pkg/routes"
)

func main() {
	r := routes.Init()

	log.Println("Server is running on port 5000..")
	err := http.ListenAndServe(":5000", r)
	if err != nil {
		log.Fatal(err)
	}
}
