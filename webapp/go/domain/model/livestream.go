package model

type LivestreamID = ID[LivestreamModel]

type LivestreamModel struct {
	ID           LivestreamID `db:"id" json:"id"`
	UserID       UserID       `db:"user_id" json:"user_id"`
	Title        string       `db:"title" json:"title"`
	Description  string       `db:"description" json:"description"`
	PlaylistUrl  string       `db:"playlist_url" json:"playlist_url"`
	ThumbnailUrl string       `db:"thumbnail_url" json:"thumbnail_url"`
	StartAt      int64        `db:"start_at" json:"start_at"`
	EndAt        int64        `db:"end_at" json:"end_at"`
}

// IsOwnedBy は userID のユーザがこのライブ配信の配信者かどうかを返す。
func (l *LivestreamModel) IsOwnedBy(userID UserID) bool {
	return l.UserID == userID
}

// Livestream は配信者・タグを含めたライブ配信の情報。
type Livestream struct {
	ID           LivestreamID
	Owner        User
	Title        string
	Description  string
	PlaylistUrl  string
	ThumbnailUrl string
	Tags         []TagModel
	StartAt      int64
	EndAt        int64
}
