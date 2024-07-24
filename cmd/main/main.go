package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/HomeArchbishop/kaption/internal/about"
	"github.com/HomeArchbishop/kaption/internal/recognizer"
	"github.com/HomeArchbishop/kaption/internal/ws"
)

func main() {
	about.PrintAbout()

	var port string
	flag.StringVar(&port, "port", "8080", "Define the server port")
	flag.Parse()

	if err := recognizer.InitModel(); err != nil {
		log.Fatal("Error during model initialization: ", err)
	}

	log.Print("Model loaded")
	log.Printf("Server started on port %s\n\n", port)

	http.HandleFunc("/socket", ws.SocketHandler)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%s", port), nil); err != nil {
		log.Fatal("Error during server start: ", err)
	}
}
