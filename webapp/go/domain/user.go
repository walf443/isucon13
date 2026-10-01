package domain

import (
	"errors"
	"slices"
)

type UserID = ID[User]

// Username はユーザ名 (ログインに使う、ユーザごとに一意な名前。表示名 DisplayName とは別)。
// 同じ string のタグ名や表示名と取り違えないように、型を分けている。
//
// ユーザ名はそのままサブドメイン (<ユーザ名>.u.isucon.dev) になるので、DNS のラベルとして使える形でなければならない。
// 型としては強制できないので、外部からの入力は ParseUsername で作ること。
// DB から読んだ値のように、すでに登録済みで信頼できるものだけ、domain.Username(s) と直接変換してよい (テストを除く)。
type Username string

// dnsLabelMaxLength は DNS のラベル (ドットで区切られた 1 つの部分) の最大長 (RFC 1035)。
const dnsLabelMaxLength = 63

// ErrInvalidUsername はユーザ名が DNS のラベルとして使えない形であることを表す。
var ErrInvalidUsername = errors.New("the username must be 1 to 63 characters of letters, digits and hyphens, and must not start or end with a hyphen")

// ParseUsername は外部からの入力を Username として読み取る。
// サブドメインとして使える形 (isDNSLabel を満たす) でなければ ErrInvalidUsername を返す。
// 大文字と小文字は変換しない。予約済みの名前かどうかは見ない (FindReservedUsername で調べる)。
func ParseUsername(s string) (Username, error) {
	if !isDNSLabel(s) {
		return "", ErrInvalidUsername
	}
	return Username(s), nil
}

// isDNSLabel は s がホスト名の 1 つのラベルとして使える形かどうかを返す (RFC 1035、RFC 1123)。
// 英数字とハイフンだけの 1〜63 文字で、先頭と末尾がハイフンでないこと。ドットは含められない (含めると別の階層のサブドメインになる)。
func isDNSLabel(s string) bool {
	if len(s) < 1 || len(s) > dnsLabelMaxLength || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		isLetterOrDigit := 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9'
		if !isLetterOrDigit && c != '-' {
			return false
		}
	}
	return true
}

// reservedUsernames はユーザ名として登録できない予約済みの名前。
var reservedUsernames = []Username{
	// 配信サービス自体のサブドメイン (pipe.u.isucon.dev) と重なる
	"pipe",
}

// FindReservedUsername は name がユーザ名として登録できない予約済みの名前かどうかを調べる。
// 予約済みの名前と完全に一致する場合だけ、その予約済みの名前 (name ではなく一覧にある値) と true を返す。
func FindReservedUsername(name Username) (Username, bool) {
	i := slices.Index(reservedUsernames, name)
	if i < 0 {
		return "", false
	}
	return reservedUsernames[i], true
}

type User struct {
	ID             UserID         `db:"id"`
	Name           Username       `db:"name"`
	DisplayName    string         `db:"display_name"`
	Description    string         `db:"description"`
	HashedPassword HashedPassword `db:"password"`
}

// UserDetail はテーマ・アイコンを含めたユーザの情報。
type UserDetail struct {
	ID          UserID
	Name        Username
	DisplayName string
	Description string
	Theme       Theme
	IconHash    IconHash
}
