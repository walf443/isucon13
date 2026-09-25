package domain

type TagID = ID[TagModel]

type TagModel struct {
	ID   TagID  `db:"id"`
	Name string `db:"name"`
}
