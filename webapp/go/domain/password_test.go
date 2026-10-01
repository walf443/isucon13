package domain

import (
	"errors"
	"fmt"
	"strings"
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

// PlainPassword は、どの書式で表示しても、パスワードの中身を出さないこと。
func TestPlainPassword_IsRedactedWhenPrinted(t *testing.T) {
	const secret = "s3cret-value"
	password := PlainPassword(secret)

	// 構造体のフィールドとして出る場合 (%+v / %#v / %v) や、スライスに入っている場合も、伏せ字になる
	type input struct {
		Name     string
		Password PlainPassword
	}
	values := map[string]string{
		"%v":            fmt.Sprintf("%v", password),
		"%s":            fmt.Sprintf("%s", password), //nolint:staticcheck // %s の書式で出力しても漏れないことを確かめるので、あえて Sprintf を使う
		"%q":            fmt.Sprintf("%q", password),
		"%+v":           fmt.Sprintf("%+v", password),
		"%#v":           fmt.Sprintf("%#v", password),
		"struct %+v":    fmt.Sprintf("%+v", input{Name: "alice", Password: password}),
		"struct %v":     fmt.Sprintf("%v", input{Name: "alice", Password: password}),
		"pointer %+v":   fmt.Sprintf("%+v", &input{Name: "alice", Password: password}),
		"slice %v":      fmt.Sprintf("%v", []PlainPassword{password}),
		"error wrapped": fmt.Errorf("login failed for %v: %w", password, errors.New("boom")).Error(),
	}
	for name, got := range values {
		if strings.Contains(got, secret) {
			t.Errorf("%s leaked the password: %q", name, got)
		}
		if !strings.Contains(got, "[REDACTED]") {
			t.Errorf("%s = %q, want it to contain [REDACTED]", name, got)
		}
	}
}

// 伏せ字にするのは表示だけで、ハッシュ化や照合には元の値が使われること。
func TestPlainPassword_RedactionDoesNotAffectHashing(t *testing.T) {
	hashed, err := HashPassword(PlainPassword("s3cret"))
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if ok, err := hashed.Matches(PlainPassword("s3cret")); err != nil || !ok {
		t.Errorf("Matches(correct) = %v, %v, want true", ok, err)
	}
	// 伏せ字そのものでは一致しない (表示用の文字列が、パスワードとして通ることはない)
	if ok, err := hashed.Matches(PlainPassword("[REDACTED]")); err != nil || ok {
		t.Errorf("Matches(redacted text) = %v, %v, want false", ok, err)
	}
}
