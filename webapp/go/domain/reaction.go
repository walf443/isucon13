package domain

type ReactionID = ID[ReactionModel]

type ReactionModel struct {
	ID           ReactionID   `db:"id"`
	EmojiName    string       `db:"emoji_name"`
	UserID       UserID       `db:"user_id"`
	LivestreamID LivestreamID `db:"livestream_id"`
	CreatedAt    int64        `db:"created_at"`
}

// ReactionDetail はリアクションしたユーザ・ライブ配信を含めたリアクションの情報。
type ReactionDetail struct {
	ID         ReactionID
	EmojiName  string
	User       UserDetail
	Livestream LivestreamDetail
	CreatedAt  int64
}
