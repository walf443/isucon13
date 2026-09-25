package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeLivecommentReportRepository はテストで設定した関数に処理を委ねる LivecommentReportRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeLivecommentReportRepository struct {
	repository.LivecommentReportRepository

	findAllWithDetailsByLivestreamID func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReport, error)
	findWithDetailsByID              func(ctx context.Context, q repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReport, error)
	create                           func(ctx context.Context, q repository.Querier, report *domain.LivecommentReportModel) (domain.LivecommentReportID, error)
	countByLivestreamID              func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error)
}

func (r *fakeLivecommentReportRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReport, error) {
	return r.findAllWithDetailsByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivecommentReportRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReport, error) {
	return r.findWithDetailsByID(ctx, q, id)
}

func (r *fakeLivecommentReportRepository) Create(ctx context.Context, q repository.Querier, report *domain.LivecommentReportModel) (domain.LivecommentReportID, error) {
	return r.create(ctx, q, report)
}

func (r *fakeLivecommentReportRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	return r.countByLivestreamID(ctx, q, livestreamID)
}
