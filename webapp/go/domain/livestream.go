package domain

type LivestreamID = ID[LivestreamModel]

// ParseLivestreamID は 10 進数の文字列をライブ配信の ID として読み取る。
func ParseLivestreamID(s string) (LivestreamID, error) {
	return ParseID[LivestreamModel](s)
}

type LivestreamModel struct {
	ID           LivestreamID `db:"id"`
	UserID       UserID       `db:"user_id"`
	Title        string       `db:"title"`
	Description  string       `db:"description"`
	PlaylistUrl  string       `db:"playlist_url"`
	ThumbnailUrl string       `db:"thumbnail_url"`
	StartAt      int64        `db:"start_at"`
	EndAt        int64        `db:"end_at"`
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
