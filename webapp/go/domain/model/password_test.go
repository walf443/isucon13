package model

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	hashed, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	// 移行前と同じく最小のコストでハッシュ化する
	if cost, err := bcrypt.Cost([]byte(hashed)); err != nil || cost != bcrypt.MinCost {
		t.Errorf("cost = %d, %v, want %d", cost, err, bcrypt.MinCost)
	}

	if ok, err := hashed.Matches("s3cret"); err != nil || !ok {
		t.Errorf("Matches(correct) = %v, %v, want true", ok, err)
	}
	if ok, err := hashed.Matches("wrong"); err != nil || ok {
		t.Errorf("Matches(wrong) = %v, %v, want false", ok, err)
	}
}

func TestHashedPassword_Matches_BrokenHash(t *testing.T) {
	// ハッシュとして不正な値は、不一致ではなくエラーにする
	ok, err := HashedPassword("broken").Matches("s3cret")
	if ok || !errors.Is(err, bcrypt.ErrHashTooShort) {
		t.Errorf("Matches = %v, %v, want false, ErrHashTooShort", ok, err)
	}
}
