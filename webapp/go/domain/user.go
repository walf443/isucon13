package domain

import "slices"

type UserID = ID[UserModel]

// reservedUsernames はユーザ名として登録できない予約済みの名前。
var reservedUsernames = []string{
	// 配信サービス自体のサブドメイン (pipe.u.isucon.dev) と重なる
	"pipe",
}

// FindReservedUsername は name がユーザ名として登録できない予約済みの名前かどうかを調べる。
// 予約済みの名前と完全に一致する場合だけ、その予約済みの名前 (name ではなく一覧にある値) と true を返す。
func FindReservedUsername(name string) (string, bool) {
	i := slices.Index(reservedUsernames, name)
	if i < 0 {
		return "", false
	}
	return reservedUsernames[i], true
}

type UserModel struct {
	ID             UserID         `db:"id"`
	Name           string         `db:"name"`
	DisplayName    string         `db:"display_name"`
	Description    string         `db:"description"`
	HashedPassword HashedPassword `db:"password"`
}

// User はテーマ・アイコンを含めたユーザの情報。
type User struct {
	ID          UserID
	Name        string
	DisplayName string
	Description string
	Theme       ThemeModel
	IconHash    IconHash
}
