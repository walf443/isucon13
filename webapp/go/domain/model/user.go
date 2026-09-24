package model

type UserModel struct {
	ID             int64  `db:"id"`
	Name           string `db:"name"`
	DisplayName    string `db:"display_name"`
	Description    string `db:"description"`
	HashedPassword string `db:"password"`
}

// User はテーマ・アイコンを含めたユーザの情報。
type User struct {
	ID          int64
	Name        string
	DisplayName string
	Description string
	Theme       ThemeModel
	IconHash    string
}
