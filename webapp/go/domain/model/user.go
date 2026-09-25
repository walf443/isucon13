package model

type UserID = ID[UserModel]

// IsReservedUsername は name がユーザ名として登録できない予約済みの名前かどうかを返す。
// "pipe" は配信サービス自体のサブドメイン (pipe.u.isucon.dev) と重なるので使えない。
func IsReservedUsername(name string) bool {
	return name == "pipe"
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
	IconHash    string
}
