package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/macihasa/parsing_httpserver/pkg/excelparsers"
	"github.com/macihasa/parsing_httpserver/pkg/xmlparsers"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Allow all origins
	CheckOrigin: func(r *http.Request) bool { return true },

	// Error handling
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Println("Upgrader error, status: ", status, reason)
	},
}

// Health is a simple health check handler
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is running.."))
}

func DCCXML(w http.ResponseWriter, r *http.Request) {
	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	log.Println("Client connected: DCC XML", conn.RemoteAddr())

	var msgchan = make(chan []byte, 1024)
	var msgcount = 0
	var finished = make(chan bool)

	go xmlparsers.DCECustomsMsg("Output", msgchan, finished)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading socket message: ", err)
		}

		msgcount++

		if msgcount%500 == 0 {
			log.Println("Files processed:", msgcount, "Active goroutines: ", runtime.NumGoroutine())
			conn.WriteMessage(websocket.TextMessage, []byte("Files processed: "+fmt.Sprint(msgcount)+"\n Active goroutines: "+fmt.Sprint(runtime.NumGoroutine())))
		}

		if string(msg) == "Finished" {
			log.Println("Client finished sending messages")
			close(msgchan)
			<-finished // Wait for the DCECustomsMsg function to finish
			log.Println("DCECustomsMsg finished")
			conn.Close()
			break
		}
		msgchan <- msg
	}
	w.WriteHeader(http.StatusOK)
	log.Println("exited loop, routine is finished.")
}

func Excel(w http.ResponseWriter, r *http.Request) {
	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}


	sheetname := mux.Vars(r)["sheet"]
	log.Println("Client request: Excel.\tSheetname:", sheetname)

	err := r.ParseMultipartForm(2 << 30) // 30=GB, 20=MB, 10=KB
	if err != nil {
		log.Println("unable to parse multipart form", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to parse multipart form"))
		return
	}

	form := r.MultipartForm

	excelparsers.CombineSheet(form, sheetname)
	log.Println("Parsing finished")

	serveCsv("./OutputXL.csv", w)
}

func HandleCsvRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Client request csv file")
	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	serveCsv("./Output.csv", w)
}

func serveCsv(filename string, w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/csv")

	filebytes, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Unable to read file", err)
	}
	w.WriteHeader(http.StatusOK)
	log.Printf("Serving file: %s\tFilesize: %v", filename, len(filebytes))
	w.Write(filebytes)
}
