package main

import (
	api "chirpy/internal"
	"log"
	"net/http"
	"sync/atomic"
)

const port = "8080"
const filepathRoot = "."

func main() {
	apiCfg := api.ApiConfig{FileserverHits: atomic.Int32{}}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.MidlewareMetricInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", api.HealthCheck)
	mux.HandleFunc("GET /api/metrics", apiCfg.RequestCount)
	mux.HandleFunc("POST /api/reset", apiCfg.Reset)
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(serv.ListenAndServe())
}
