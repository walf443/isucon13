package domain

type ThemeID = ID[Theme]

type Theme struct {
	ID       ThemeID `db:"id"`
	UserID   UserID  `db:"user_id"`
	DarkMode bool    `db:"dark_mode"`
}
