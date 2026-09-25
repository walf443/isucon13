package model

type ReactionID = ID[ReactionModel]

type ReactionModel struct {
	ID           ReactionID   `db:"id"`
	EmojiName    string       `db:"emoji_name"`
	UserID       UserID       `db:"user_id"`
	LivestreamID LivestreamID `db:"livestream_id"`
	CreatedAt    int64        `db:"created_at"`
}

// Reaction はリアクションしたユーザ・ライブ配信を含めたリアクションの情報。
type Reaction struct {
	ID         ReactionID
	EmojiName  string
	User       User
	Livestream Livestream
	CreatedAt  int64
}
