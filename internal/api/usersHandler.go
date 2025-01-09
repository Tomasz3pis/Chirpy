package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
)

func (cfg *ApiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(cfg.Platform) != "dev" {
		respondWithError(w, 403, "Not allowed. Only for testing purpous")
		return
	}
	err := cfg.Db.DeteleUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	cfg.FileserverHits = atomic.Int32{}
	respondWithJSON(w, 200, "Succesfully deleted all users")
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	type email struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	e := email{}
	err := decoder.Decode(&e)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	u, err := cfg.Db.CreateUser(r.Context(), e.Email)
	if err != nil {
		respondWithError(w, 500, "Failed to create user")
		return
	}

	respondWithJSON(w, 201, user{
		Id:         u.ID.String(),
		Created_at: u.CreatedAt,
		Updated_at: u.UpdatedAt,
		Email:      u.Email,
	})

}
