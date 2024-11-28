package main

import (
	handlers "chirpy/internal"
	"log"
	"net/http"
)

const port = "8080"
const filepathRoot = "."

func main() {

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.HandleFunc("/healthz/", handlers.HealthCheck)

	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(serv.ListenAndServe())
}
