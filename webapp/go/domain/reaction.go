package domain

type ReactionID = ID[Reaction]

type Reaction struct {
	ID           ReactionID   `db:"id"`
	EmojiName    string       `db:"emoji_name"`
	UserID       UserID       `db:"user_id"`
	LivestreamID LivestreamID `db:"livestream_id"`
	CreatedAt    int64        `db:"created_at"`
}

// PrimaryKey はエンティティを一意に識別する ID を返す。ID ごとの map に詰めたものを元の順序に並べ直すときなどに、ジェネリクスから ID を取り出すために使う。
func (r *Reaction) PrimaryKey() ReactionID {
	return r.ID
}

// ReactionDetail はリアクションしたユーザ・ライブ配信を含めたリアクションの情報。
type ReactionDetail struct {
	ID         ReactionID
	EmojiName  string
	User       UserDetail
	Livestream LivestreamDetail
	CreatedAt  int64
}
