package domain

type LivestreamTagID = ID[LivestreamTag]

// LivestreamTag はライブ配信とタグの紐付け。
type LivestreamTag struct {
	ID           LivestreamTagID `db:"id"`
	LivestreamID LivestreamID    `db:"livestream_id"`
	TagID        TagID           `db:"tag_id"`
}
