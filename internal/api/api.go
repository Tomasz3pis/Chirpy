package api

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Db             *database.Queries
	Platform       string
}

type user struct {
	Id         string    `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Email      string    `json:"email"`
}

type chirp struct {
	Id         string    `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    string    `json:"user_id"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func (cfg *ApiConfig) MidlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) RequestCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	tmp := `
		<html>
		  <body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		  </body>
		</html>`
	res := fmt.Sprintf(tmp, cfg.FileserverHits.Load())

	w.Write([]byte(res))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
