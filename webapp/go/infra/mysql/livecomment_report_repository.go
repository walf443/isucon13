package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type livecommentReportRepository struct {
	// defaultIconHash はアイコン未登録のユーザに使う既定のアイコンのハッシュ。
	defaultIconHash domain.IconHash
}

func NewLivecommentReportRepository(defaultIconHash domain.IconHash) repository.LivecommentReportRepository {
	return &livecommentReportRepository{defaultIconHash: defaultIconHash}
}

func (r *livecommentReportRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReport, error) {
	var reportModels []*domain.LivecommentReportModel
	if err := q.SelectContext(ctx, &reportModels, "SELECT id, user_id, livestream_id, livecomment_id, created_at FROM livecomment_reports WHERE livestream_id = ?", livestreamID); err != nil {
		return nil, err
	}
	return fillLivecommentReports(ctx, q, reportModels, r.defaultIconHash)
}

func (r *livecommentReportRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReport, error) {
	var reportModel domain.LivecommentReportModel
	err := q.GetContext(ctx, &reportModel, "SELECT id, user_id, livestream_id, livecomment_id, created_at FROM livecomment_reports WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	reports, err := fillLivecommentReports(ctx, q, []*domain.LivecommentReportModel{&reportModel}, r.defaultIconHash)
	if err != nil {
		return nil, err
	}
	return reports[0], nil
}

func (r *livecommentReportRepository) Create(ctx context.Context, q repository.Querier, report *domain.LivecommentReportModel) (domain.LivecommentReportID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)", report.UserID, report.LivestreamID, report.LivecommentID, report.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return domain.LivecommentReportID(id), nil
}

func (r *livecommentReportRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var totalReports int64
	if err := q.GetContext(ctx, &totalReports, `SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil {
		return 0, err
	}
	return totalReports, nil
}
