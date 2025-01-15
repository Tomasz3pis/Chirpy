package main

import (
	"chirpy/internal/api"
	"chirpy/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const port = "8080"
const filepathRoot = "."

func main() {
	godotenv.Load()

	apiCfg := api.ApiConfig{
		FileserverHits: atomic.Int32{},
		Db:             getDb(),
		Platform:       os.Getenv("PLATFORM"),
		JWT:            os.Getenv("JWT_SECRET"),
		Polka_key:      os.Getenv("POLKA_KEY"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.MidlewareMetricInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", api.HealthCheck)
	mux.HandleFunc("GET /admin/metrics", apiCfg.RequestCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.Reset)
	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetAllChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.GetChirp)
	mux.HandleFunc("DELETE /api/chirps/{id}", apiCfg.DeleteChirp)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.Login)
	mux.HandleFunc("POST /api/refresh", apiCfg.Refresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.Revoke)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeUser)
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(serv.ListenAndServe())
}

func getDb() *database.Queries {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Failed to connect to db: %s\n", err)
	}
	return database.New(db)
}
