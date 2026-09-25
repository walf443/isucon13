package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type livestreamViewersHistoryRepository struct{}

func NewLivestreamViewersHistoryRepository() repository.LivestreamViewersHistoryRepository {
	return &livestreamViewersHistoryRepository{}
}

func (r *livestreamViewersHistoryRepository) Create(ctx context.Context, q repository.Querier, viewer *model.LivestreamViewersHistoryModel) error {
	_, err := q.ExecContext(ctx, "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(?, ?, ?)", viewer.UserID, viewer.LivestreamID, viewer.CreatedAt)
	return err
}

func (r *livestreamViewersHistoryRepository) DeleteByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) error {
	_, err := q.ExecContext(ctx, "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?", userID, livestreamID)
	return err
}

func (r *livestreamViewersHistoryRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	var cnt int64
	if err := q.GetContext(ctx, &cnt, "SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?", livestreamID); err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *livestreamViewersHistoryRepository) CountViewersByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	var viewersCount int64
	if err := q.GetContext(ctx, &viewersCount, `SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil {
		return 0, err
	}
	return viewersCount, nil
}
