package domain

type NGWordID = ID[NGWord]

type NGWord struct {
	ID           NGWordID     `db:"id"`
	UserID       UserID       `db:"user_id"`
	LivestreamID LivestreamID `db:"livestream_id"`
	Word         string       `db:"word"`
	CreatedAt    int64        `db:"created_at"`
}
