package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeLivecommentRepository はテストで設定した関数に処理を委ねる LivecommentRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeLivecommentRepository struct {
	repository.LivecommentRepository

	findByID                                func(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.LivecommentModel, error)
	findWithDetailsByID                     func(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.Livecomment, error)
	findAllWithDetailsByLivestreamID        func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error)
	findAllWithDetailsByLivestreamIDLimited func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit model.Limit) ([]*model.Livecomment, error)
	findAllByLivestreamID                   func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentModel, error)
	create                                  func(ctx context.Context, q repository.Querier, livecomment *model.LivecommentModel) (model.LivecommentID, error)
	deleteAllByLivestreamIDMatchingNGWord   func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, word string) error
	sumTip                                  func(ctx context.Context, q repository.Querier) (int64, error)
	sumTipByLivestreamID                    func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error)
	sumTipByLivestreamOwnerID               func(ctx context.Context, q repository.Querier, userID model.UserID) (int64, error)
	maxTipByLivestreamID                    func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error)
}

func (r *fakeLivecommentRepository) FindByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.LivecommentModel, error) {
	return r.findByID(ctx, q, id)
}

func (r *fakeLivecommentRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.Livecomment, error) {
	return r.findWithDetailsByID(ctx, q, id)
}

func (r *fakeLivecommentRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error) {
	return r.findAllWithDetailsByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivecommentRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit model.Limit) ([]*model.Livecomment, error) {
	return r.findAllWithDetailsByLivestreamIDLimited(ctx, q, livestreamID, limit)
}

func (r *fakeLivecommentRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentModel, error) {
	return r.findAllByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivecommentRepository) Create(ctx context.Context, q repository.Querier, livecomment *model.LivecommentModel) (model.LivecommentID, error) {
	return r.create(ctx, q, livecomment)
}

func (r *fakeLivecommentRepository) DeleteAllByLivestreamIDMatchingNGWord(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, word string) error {
	return r.deleteAllByLivestreamIDMatchingNGWord(ctx, q, livestreamID, word)
}

func (r *fakeLivecommentRepository) SumTip(ctx context.Context, q repository.Querier) (int64, error) {
	return r.sumTip(ctx, q)
}

func (r *fakeLivecommentRepository) SumTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.sumTipByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivecommentRepository) SumTipByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID model.UserID) (int64, error) {
	return r.sumTipByLivestreamOwnerID(ctx, q, userID)
}

func (r *fakeLivecommentRepository) MaxTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.maxTipByLivestreamID(ctx, q, livestreamID)
}
