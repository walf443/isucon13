package domain

type LivestreamViewersHistoryID = ID[LivestreamViewersHistoryModel]

// LivestreamViewersHistoryModel はライブ配信の視聴履歴 (livestream_viewers_history) の1行。
type LivestreamViewersHistoryModel struct {
	ID           LivestreamViewersHistoryID `db:"id"`
	UserID       UserID                     `db:"user_id"`
	LivestreamID LivestreamID               `db:"livestream_id"`
	CreatedAt    int64                      `db:"created_at"`
}
