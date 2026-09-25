package model

type UserID = ID[UserModel]

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
