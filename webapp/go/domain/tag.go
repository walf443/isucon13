package domain

type TagID = ID[Tag]

type Tag struct {
	ID   TagID  `db:"id"`
	Name string `db:"name"`
}
