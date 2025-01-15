package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) (string, error) {
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(pw), 10)

	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		return "", err
	}
	return string(hashedPw[:]), nil

}

func CheckPasswordHash(pw, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	if err != nil {
		log.Printf("Failed to compare hash password: %s", err)
		return err
	}
	return nil
}

func GetAPIKey(headers http.Header) (string, error) {
	key := headers.Get("Authorization")
	if key == "" {
		return "", fmt.Errorf("Key not found")
	}
	return strings.Split(key, "ApiKey ")[1], nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := headers.Get("Authorization")
	if tokenString == "" {
		return "", fmt.Errorf("Token not found")
	}
	return strings.Split(tokenString, "Bearer ")[1], nil
}

func MakeJWT(userId uuid.UUID, tokenSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Duration(60*60) * time.Second)},
		Subject:   userId.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	userId, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, err
	}
	return userId, nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
