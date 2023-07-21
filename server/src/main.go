package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/macihasa/parsing_httpserver/pkg/routes"
)

func main() {
	r := routes.Init()

	// Profiler:
	go func ()  {
		log.Println("Profiler is running on port 6060..")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Println("Server is running on port 5000..")
	err := http.ListenAndServe("localhost:5000", r)
	if err != nil {
		log.Fatal(err)
	}
}
