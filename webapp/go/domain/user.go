package domain

import "slices"

type UserID = ID[User]

// Username はユーザ名 (ログインに使う、ユーザごとに一意な名前。表示名 DisplayName とは別)。
// 同じ string のタグ名や表示名と取り違えないように、型を分けている。
type Username string

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
