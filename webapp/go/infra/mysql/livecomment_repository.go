package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type livecommentRepository struct {
	// defaultIconHash はアイコン未登録のユーザに使う既定のアイコンのハッシュ。
	defaultIconHash model.IconHash
}

func NewLivecommentRepository(defaultIconHash model.IconHash) repository.LivecommentRepository {
	return &livecommentRepository{defaultIconHash: defaultIconHash}
}

func (r *livecommentRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error) {
	var livecommentModels []*model.LivecommentModel
	if err := q.SelectContext(ctx, &livecommentModels, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC", livestreamID); err != nil {
		return nil, err
	}
	return fillLivecomments(ctx, q, livecommentModels, r.defaultIconHash)
}

func (r *livecommentRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit model.Limit) ([]*model.Livecomment, error) {
	var livecommentModels []*model.LivecommentModel
	if err := q.SelectContext(ctx, &livecommentModels, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?", livestreamID, limit); err != nil {
		return nil, err
	}
	return fillLivecomments(ctx, q, livecommentModels, r.defaultIconHash)
}

func (r *livecommentRepository) FindByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.LivecommentModel, error) {
	var livecommentModel model.LivecommentModel
	err := q.GetContext(ctx, &livecommentModel, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &livecommentModel, nil
}

func (r *livecommentRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.Livecomment, error) {
	var livecommentModel model.LivecommentModel
	err := q.GetContext(ctx, &livecommentModel, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	livecomments, err := fillLivecomments(ctx, q, []*model.LivecommentModel{&livecommentModel}, r.defaultIconHash)
	if err != nil {
		return nil, err
	}
	return livecomments[0], nil
}

func (r *livecommentRepository) Create(ctx context.Context, q repository.Querier, livecomment *model.LivecommentModel) (model.LivecommentID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)", livecomment.UserID, livecomment.LivestreamID, livecomment.Comment, livecomment.Tip, livecomment.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return model.LivecommentID(id), nil
}

func (r *livecommentRepository) DeleteAllByLivestreamIDMatchingNGWord(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, word string) error {
	// 移行前と同じクエリの流れ (全ライブコメントを取得し、1件ずつ条件付きで DELETE する) を保っている
	var livecomments []*model.LivecommentModel
	if err := q.SelectContext(ctx, &livecomments, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments"); err != nil {
		return fmt.Errorf("failed to get livecomments: %w", err)
	}

	for _, livecomment := range livecomments {
		query := `
			DELETE FROM livecomments
			WHERE
			id = ? AND
			livestream_id = ? AND
			(SELECT COUNT(*)
			FROM
			(SELECT ? AS text) AS texts
			INNER JOIN
			(SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
			ON texts.text LIKE patterns.pattern) >= 1;
			`
		if _, err := q.ExecContext(ctx, query, livecomment.ID, livestreamID, livecomment.Comment, word); err != nil {
			return fmt.Errorf("failed to delete old livecomments that hit spams: %w", err)
		}
	}
	return nil
}

func (r *livecommentRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentModel, error) {
	var livecomments []*model.LivecommentModel
	if err := q.SelectContext(ctx, &livecomments, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE livestream_id = ?", livestreamID); err != nil {
		return nil, err
	}
	return livecomments, nil
}

func (r *livecommentRepository) SumTipByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID model.UserID) (int64, error) {
	var tips int64
	// クエリ文字列は移行前のまま (空白も含めて) にしている
	query := `
		SELECT IFNULL(SUM(l2.tip), 0) FROM users u
		INNER JOIN livestreams l ON l.user_id = u.id	
		INNER JOIN livecomments l2 ON l2.livestream_id = l.id
		WHERE u.id = ?`
	if err := q.GetContext(ctx, &tips, query, userID); err != nil {
		return 0, err
	}
	return tips, nil
}

func (r *livecommentRepository) SumTip(ctx context.Context, q repository.Querier) (int64, error) {
	var totalTip int64
	if err := q.GetContext(ctx, &totalTip, "SELECT IFNULL(SUM(tip), 0) FROM livecomments"); err != nil {
		return 0, err
	}
	return totalTip, nil
}

func (r *livecommentRepository) SumTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	var totalTips int64
	if err := q.GetContext(ctx, &totalTips, "SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?", livestreamID); err != nil {
		return 0, err
	}
	return totalTips, nil
}

func (r *livecommentRepository) MaxTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	var maxTip int64
	if err := q.GetContext(ctx, &maxTip, `SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil {
		return 0, err
	}
	return maxTip, nil
}
