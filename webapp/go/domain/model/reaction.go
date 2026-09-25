package model

type ReactionModel struct {
	ID           int64  `db:"id"`
	EmojiName    string `db:"emoji_name"`
	UserID       int64  `db:"user_id"`
	LivestreamID int64  `db:"livestream_id"`
	CreatedAt    int64  `db:"created_at"`
}

// Reaction はリアクションしたユーザ・ライブ配信を含めたリアクションの情報。
type Reaction struct {
	ID         int64
	EmojiName  string
	User       User
	Livestream Livestream
	CreatedAt  int64
}
