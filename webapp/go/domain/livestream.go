package domain

type LivestreamID = ID[Livestream]

// ParseLivestreamID は 10 進数の文字列をライブ配信の ID として読み取る。
func ParseLivestreamID(s string) (LivestreamID, error) {
	return ParseID[Livestream](s)
}

type Livestream struct {
	ID           LivestreamID `db:"id"`
	UserID       UserID       `db:"user_id"`
	Title        string       `db:"title"`
	Description  string       `db:"description"`
	PlaylistUrl  string       `db:"playlist_url"`
	ThumbnailUrl string       `db:"thumbnail_url"`
	StartAt      int64        `db:"start_at"`
	EndAt        int64        `db:"end_at"`
}

// PrimaryKey はエンティティを一意に識別する ID を返す。ID ごとの map に詰めたものを元の順序に並べ直すときなどに、ジェネリクスから ID を取り出すために使う。
func (l *Livestream) PrimaryKey() LivestreamID {
	return l.ID
}

// IsOwnedBy は userID のユーザがこのライブ配信の配信者かどうかを返す。
func (l *Livestream) IsOwnedBy(userID UserID) bool {
	return l.UserID == userID
}

// LivestreamDetail は配信者・タグを含めたライブ配信の情報。
type LivestreamDetail struct {
	ID           LivestreamID
	Owner        UserDetail
	Title        string
	Description  string
	PlaylistUrl  string
	ThumbnailUrl string
	Tags         []Tag
	StartAt      int64
	EndAt        int64
}
