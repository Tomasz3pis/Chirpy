package api

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Db             *database.Queries
	Platform       string
	JWT            string
	Polka_key      string
}

type user struct {
	Id            string    `json:"id"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
	Email         string    `json:"email"`
	Token         string    `json:"token"`
	Refresh_token string    `json:"refresh_token"`
	Is_chirpy_red bool      `json:"is_chirpy_red"`
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

func (cfg *ApiConfig) Refresh(w http.ResponseWriter, r *http.Request) {
	type tokenResp struct {
		Token string `json:"token"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to refresh token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	rt, err := cfg.Db.GetToken(r.Context(), token)
	if err != nil || rt.ExpiresAt.Before(time.Now()) {
		log.Printf("Failed to fetch token: %s", err)
		respondWithError(w, 401, "Token not valid")
		return
	}
	if rt.RevokedAt != (sql.NullTime{}) {
		respondWithError(w, 401, "Token not valid")
		return
	}

	token, err = auth.MakeJWT(rt.UserID, cfg.JWT)
	if err != nil {
		log.Printf("Failed to create token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 200, tokenResp{Token: token})
}

func (cfg *ApiConfig) Revoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to refresh token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	err = cfg.Db.RevokeToken(r.Context(), token)
	if err != nil {
		log.Printf("Failed to revoke token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return

	}
	w.WriteHeader(204)
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
