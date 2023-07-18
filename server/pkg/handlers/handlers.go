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
		log.Println(reason)
	},
}

// Health is a simple health check handler
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is running.."))
}

func DCCXML(w http.ResponseWriter, r *http.Request) {
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
}

func Excel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	sheetname := mux.Vars(r)["sheet"]
	log.Println("Client request: Excel.\tSheetname:", sheetname)

	err := r.ParseMultipartForm(2 << 30)
	if err != nil {
		log.Println("unable to parse multipart form", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to parse multipart form"))
		return
	}

	rowsch := make(chan []string, 1024)
	finished := make(chan bool)

	form := r.MultipartForm

	go excelparsers.CombineSheet(form, rowsch, sheetname, finished)

	<-finished
	log.Println("Parsing finished")

	filebytes, err := os.ReadFile("./OutputXL.csv")
	if err != nil {
		log.Println(err)
	}
	log.Println("File size: ", len(filebytes))

	w.Header().Set("Content-Type", "text/csv")
	w.Write(filebytes)
	log.Println("File served..")

}

func ServeCSVfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Client request: ServeCSVfile")
	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	filebytes, err := os.ReadFile("./Output.csv")
	if err != nil {
		log.Println(err)
	}
	log.Println("File size: ", len(filebytes))

	w.Header().Set("Content-Type", "text/csv")
	w.Write(filebytes)
	log.Println("File served..")
}
