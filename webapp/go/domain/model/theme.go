package model

type ThemeID = ID[ThemeModel]

type ThemeModel struct {
	ID       ThemeID `db:"id"`
	UserID   UserID  `db:"user_id"`
	DarkMode bool    `db:"dark_mode"`
}
