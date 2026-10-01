package domain

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// passwordHashCost は bcrypt のコスト。移行前と同じく最小のコストにしている。
const passwordHashCost = bcrypt.MinCost

// PlainPassword はハッシュ化する前のパスワード。
// ログなどにうっかり出さないように、文字列として表示する (%v / %s / %q / %+v / %#v) と伏せ字になる。
// 中身が必要な場合は、HashPassword / Matches に渡すか、string(p) と明示的に変換する。
type PlainPassword string

// redactedPassword は PlainPassword を表示するときの伏せ字。
const redactedPassword = "[REDACTED]"

func (PlainPassword) String() string { return redactedPassword }

func (PlainPassword) GoString() string { return redactedPassword }

// HashedPassword は bcrypt でハッシュ化したパスワード。
type HashedPassword string

// HashPassword は平文のパスワードをハッシュ化する。
func HashPassword(plain PlainPassword) (HashedPassword, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), passwordHashCost)
	if err != nil {
		return "", err
	}
	return HashedPassword(hashed), nil
}

// Matches は plain がこのハッシュのパスワードと一致するかを返す。
// 一致しない場合は false を返し、ハッシュが壊れている場合などはエラーを返す。
func (h HashedPassword) Matches(plain PlainPassword) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
