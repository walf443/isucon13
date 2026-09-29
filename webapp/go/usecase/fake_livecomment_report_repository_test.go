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

	findByID              func(ctx context.Context, q repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReportModel, error)
	findAllByLivestreamID func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReportModel, error)
	create                func(ctx context.Context, q repository.Querier, report *domain.LivecommentReportModel) (domain.LivecommentReportID, error)
	countByLivestreamID   func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error)
}

func (r *fakeLivecommentReportRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReportModel, error) {
	return r.findByID(ctx, q, id)
}

func (r *fakeLivecommentReportRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReportModel, error) {
	return r.findAllByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivecommentReportRepository) Create(ctx context.Context, q repository.Querier, report *domain.LivecommentReportModel) (domain.LivecommentReportID, error) {
	return r.create(ctx, q, report)
}

func (r *fakeLivecommentReportRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	return r.countByLivestreamID(ctx, q, livestreamID)
}
