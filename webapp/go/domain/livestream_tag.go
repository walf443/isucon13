package domain

type LivestreamTagID = ID[LivestreamTagModel]

// LivestreamTagModel はライブ配信とタグの紐付け。
type LivestreamTagModel struct {
	ID           LivestreamTagID `db:"id"`
	LivestreamID LivestreamID    `db:"livestream_id"`
	TagID        TagID           `db:"tag_id"`
}
