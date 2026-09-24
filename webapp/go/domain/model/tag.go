package model

type TagModel struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}
