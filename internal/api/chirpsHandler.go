package api

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
)

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.PathValue("id"))
	cId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	c, err := cfg.Db.GetChirp(r.Context(), cId)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, chirp{
		Id:         c.ID.String(),
		Created_at: c.CreatedAt,
		Updated_at: c.UpdatedAt,
		Body:       c.Body,
		User_id:    c.UserID.String(),
	})
}

func (cfg *ApiConfig) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	cs, err := cfg.Db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	var respC []chirp
	for _, c := range cs {
		respC = append(respC, chirp{
			Id:         c.ID.String(),
			Created_at: c.CreatedAt,
			Updated_at: c.UpdatedAt,
			Body:       c.Body,
			User_id:    c.UserID.String(),
		})
	}
	respondWithJSON(w, 200, respC)
}

func (cfg *ApiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body    string `json:"body"`
		User_id string `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	uId, err := uuid.Parse(params.User_id)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	c, err := cfg.Db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: uId,
	})
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 201, chirp{
		Id:         c.ID.String(),
		Created_at: c.CreatedAt,
		Updated_at: c.UpdatedAt,
		Body:       c.Body,
		User_id:    c.UserID.String(),
	})
}

func censorship(s string) string {
	profane := []string{"kerfuffle", "sharbert", "fornax"}
	var clean []string
	for _, v := range strings.Fields(s) {
		if slices.Contains(profane, strings.ToLower(v)) {
			v = "****"
		}
		clean = append(clean, v)
	}
	return strings.Join(clean, " ")
}
