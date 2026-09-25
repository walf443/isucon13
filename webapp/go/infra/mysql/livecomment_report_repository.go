package mysql

import (
	"context"

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
