package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type livecommentReportRepository struct {
	// fallbackIcon はアイコン未登録のユーザに使う画像。
	fallbackIcon []byte
}

func NewLivecommentReportRepository(fallbackIcon []byte) repository.LivecommentReportRepository {
	return &livecommentReportRepository{fallbackIcon: fallbackIcon}
}

func (r *livecommentReportRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
	var reportModels []*model.LivecommentReportModel
	if err := q.SelectContext(ctx, &reportModels, "SELECT id, user_id, livestream_id, livecomment_id, created_at FROM livecomment_reports WHERE livestream_id = ?", livestreamID); err != nil {
		return nil, err
	}
	return fillLivecommentReports(ctx, q, reportModels, r.fallbackIcon)
}

func (r *livecommentReportRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivecommentReportID) (*model.LivecommentReport, error) {
	var reportModel model.LivecommentReportModel
	err := q.GetContext(ctx, &reportModel, "SELECT id, user_id, livestream_id, livecomment_id, created_at FROM livecomment_reports WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	reports, err := fillLivecommentReports(ctx, q, []*model.LivecommentReportModel{&reportModel}, r.fallbackIcon)
	if err != nil {
		return nil, err
	}
	return reports[0], nil
}

func (r *livecommentReportRepository) Create(ctx context.Context, q repository.Querier, report *model.LivecommentReportModel) (model.LivecommentReportID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)", report.UserID, report.LivestreamID, report.LivecommentID, report.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return model.LivecommentReportID(id), nil
}

func (r *livecommentReportRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	var totalReports int64
	if err := q.GetContext(ctx, &totalReports, `SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil {
		return 0, err
	}
	return totalReports, nil
}
