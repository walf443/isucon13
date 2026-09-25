package domain

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// passwordHashCost は bcrypt のコスト。移行前と同じく最小のコストにしている。
const passwordHashCost = bcrypt.MinCost

// HashedPassword は bcrypt でハッシュ化したパスワード。
type HashedPassword string

// HashPassword は平文のパスワードをハッシュ化する。
func HashPassword(plain string) (HashedPassword, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), passwordHashCost)
	if err != nil {
		return "", err
	}
	return HashedPassword(hashed), nil
}

// Matches は plain がこのハッシュのパスワードと一致するかを返す。
// 一致しない場合は false を返し、ハッシュが壊れている場合などはエラーを返す。
func (h HashedPassword) Matches(plain string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
