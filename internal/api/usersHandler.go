package api

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type input struct {
	Email string `json:"email"`
	Pw    string `json:"password"`
}

func (cfg *ApiConfig) Login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	e := input{}
	err := decoder.Decode(&e)
	if err != nil {
		log.Printf("Failed to decode json: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	u, err := cfg.Db.GetUserByEmail(r.Context(), e.Email)
	if err != nil {
		log.Printf("Failed to fetch user from db: %s", err)
		respondWithError(w, 401, "Incorrect email or password.")
		return
	}
	err = auth.CheckPasswordHash(e.Pw, u.HashedPassword)
	if err != nil {
		log.Printf("Failed to verify password: %s", err)
		respondWithError(w, 401, "Incorrect email or password.")
		return
	}
	token, err := auth.MakeJWT(u.ID, cfg.JWT)
	if err != nil {
		log.Printf("Failed to create JWT: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	rt, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Failed to create refresh token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	_, err = cfg.Db.CreateToken(r.Context(), database.CreateTokenParams{
		Token:     rt,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(time.Duration(60*60*24*60) * time.Second),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		log.Printf("Failed to save refresh token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 200, user{
		Id:            u.ID.String(),
		Created_at:    u.CreatedAt,
		Updated_at:    u.UpdatedAt,
		Email:         u.Email,
		Token:         token,
		Refresh_token: rt,
		Is_chirpy_red: u.IsChirpyRed,
	})
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	e := input{}
	err := decoder.Decode(&e)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	pwHash, err := auth.HashPassword(e.Pw)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}
	u, err := cfg.Db.CreateUser(r.Context(), database.CreateUserParams{Email: e.Email, HashedPassword: pwHash})
	if err != nil {
		log.Printf("Error during creating user: %s", err)
		respondWithError(w, 500, "Failed to create user")
		return
	}
	respondWithJSON(w, 201, user{
		Id:            u.ID.String(),
		Created_at:    u.CreatedAt,
		Updated_at:    u.UpdatedAt,
		Email:         u.Email,
		Is_chirpy_red: u.IsChirpyRed,
	})
}

func (cfg *ApiConfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to read token: %s", err)
		respondWithError(w, 401, "Invalid token")
		return
	}
	uId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		log.Printf("Failed to validate token: %s", err)
		respondWithError(w, 401, "Invalid token")
		return
	}
	decoder := json.NewDecoder(r.Body)
	e := input{}
	err = decoder.Decode(&e)
	if err != nil {
		log.Printf("Failed to decode json: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	hashPw, err := auth.HashPassword(e.Pw)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	u, err := cfg.Db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          e.Email,
		HashedPassword: hashPw,
		ID:             uId,
	})
	if err != nil {
		log.Printf("Failed to update user: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	respondWithJSON(w, 200, user{
		Id:            u.ID.String(),
		Created_at:    u.CreatedAt,
		Updated_at:    u.UpdatedAt,
		Email:         u.Email,
		Is_chirpy_red: u.IsChirpyRed,
	})
}

func (cfg *ApiConfig) UpgradeUser(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	key, err := auth.GetAPIKey(r.Header)
	if err != nil || key != cfg.Polka_key {
		log.Printf("Failed to read api key: %s", err)
		respondWithError(w, 401, "Invalid api key")
		return
	}
	decoder := json.NewDecoder(r.Body)
	e := input{}
	err = decoder.Decode(&e)
	if err != nil {
		log.Printf("Failed to decode json: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if e.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	uId, err := uuid.Parse(e.Data.UserId)
	if err != nil {
		log.Printf("Failed to parse user id: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	_, err = cfg.Db.UpdateUserRed(r.Context(), uId)
	if err != nil {
		log.Printf("Failed to upgrade user id: %s", err)
		respondWithError(w, 404, "User not found")
		return
	}
	w.WriteHeader(204)
}

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
