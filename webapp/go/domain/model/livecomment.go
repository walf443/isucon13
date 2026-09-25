package model

type LivecommentID = ID[LivecommentModel]

type LivecommentModel struct {
	ID           LivecommentID `db:"id"`
	UserID       UserID        `db:"user_id"`
	LivestreamID LivestreamID  `db:"livestream_id"`
	Comment      string        `db:"comment"`
	Tip          int64         `db:"tip"`
	CreatedAt    int64         `db:"created_at"`
}
