package api

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
)

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
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
	s := r.URL.Query().Get("author_id")
	var cs []database.Chirp
	var err error
	if s != "" {
		uId, _ := uuid.Parse(s)
		cs, err = cfg.Db.GetAllChirpsByAuthor(r.Context(), uId)
	} else {
		cs, err = cfg.Db.GetAllChirps(r.Context())
	}
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
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to retrive token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	uId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		respondWithError(w, 401, "Failed to validate user")
		return
	}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Failed to decode json: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	c, err := cfg.Db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: uId,
	})
	if err != nil {
		log.Printf("Failed to create chirp: %s", err)
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

func (cfg *ApiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	cId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to retrive token: %s", err)
		respondWithError(w, 401, "Invalid token")
		return
	}
	uId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		log.Printf("Failed to validate token: %s", err)
		respondWithError(w, 403, "Invalid token")
		return
	}
	chirpDb, err := cfg.Db.GetChirp(r.Context(), cId)
	if err != nil {
		log.Printf("Failed to retrive chirp: %s", err)
		respondWithError(w, 404, "Chirp not found")
		return
	}
	if chirpDb.UserID != uId {
		respondWithError(w, 403, "Not authorized")
		return
	}
	err = cfg.Db.DeleteChirp(r.Context(), cId)
	if err != nil {
		log.Printf("Failed to delete chirp: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	w.WriteHeader(204)

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
