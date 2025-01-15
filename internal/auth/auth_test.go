package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	pw := "password"
	hash, err := HashPassword(pw)
	if err != nil {
		t.Errorf("Hashing failed: %s", err)
	}
	err = CheckPasswordHash(pw, hash)
	if err != nil {
		t.Errorf("Hash and password missmatch: %s", err)
	}
}

func TestJWT(t *testing.T) {

	userId, err := uuid.NewUUID()
	if err != nil {
		t.Errorf("Failed to generate UUID: %s", err)
	}

	var tests = []struct {
		s   string
		uId uuid.UUID
		w   uuid.UUID
	}{
		{"correct", userId, userId},
		{"correct", userId, uuid.UUID{}},
		{"incorrect", userId, uuid.UUID{}},
	}
	for _, tt := range tests {
		token, err := MakeJWT(tt.uId, tt.s)
		if err != nil {
			t.Errorf("Failed to create token: %s", err)
		}
		uid, err := ValidateJWT(token, "correct")
		if uid != tt.w {
			t.Errorf("Wanted %s, got %s", tt.w, uid)
		}
	}

}
