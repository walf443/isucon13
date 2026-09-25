package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeLivecommentRepository struct {
	livecommentModel *model.LivecommentModel
	findErr          error
	// livecommentModelsByLivestreamID は FindAllByLivestreamID が返すライブコメント
	livecommentModelsByLivestreamID map[model.LivestreamID][]*model.LivecommentModel
	findAllByLivestreamIDErr        error
	// tipsByOwnerID は SumTipByLivestreamOwnerID が返すチップ合計
	tipsByOwnerID map[model.UserID]int64
	sumTipErr     error
	// totalTip は SumTip が返すチップ合計
	totalTip    int64
	totalTipErr error
	// tipsByLivestreamID は SumTipByLivestreamID が返すチップ合計
	tipsByLivestreamID    map[model.LivestreamID]int64
	sumTipByLivestreamErr error
	maxTip                int64
	maxTipErr             error
	gotMaxTipLivestreamID model.LivestreamID

	deleteErr error
	// deletedWords は DeleteAllByLivestreamIDMatchingNGWord に渡された NG ワード
	deletedWords []string

	livecomment  *model.Livecomment
	livecomments []*model.Livecomment
	err          error
	createID     model.LivecommentID
	createErr    error
	gotID        model.LivecommentID
	gotCreated   *model.LivecommentModel

	calls           []string
	gotLivestreamID model.LivestreamID
	gotLimit        model.Limit
}

func (r *fakeLivecommentRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentModel, error) {
	return r.livecommentModelsByLivestreamID[livestreamID], r.findAllByLivestreamIDErr
}

func (r *fakeLivecommentRepository) SumTip(ctx context.Context, q repository.Querier) (int64, error) {
	return r.totalTip, r.totalTipErr
}

func (r *fakeLivecommentRepository) SumTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.tipsByLivestreamID[livestreamID], r.sumTipByLivestreamErr
}

func (r *fakeLivecommentRepository) MaxTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	r.gotMaxTipLivestreamID = livestreamID
	return r.maxTip, r.maxTipErr
}

func (r *fakeLivecommentRepository) SumTipByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID model.UserID) (int64, error) {
	return r.tipsByOwnerID[userID], r.sumTipErr
}

func (r *fakeLivecommentRepository) FindByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.LivecommentModel, error) {
	r.calls = append(r.calls, "FindByID")
	r.gotID = id
	return r.livecommentModel, r.findErr
}

func (r *fakeLivecommentRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.Livecomment, error) {
	r.calls = append(r.calls, "FindWithDetailsByID")
	r.gotID = id
	return r.livecomment, r.err
}

func (r *fakeLivecommentRepository) DeleteAllByLivestreamIDMatchingNGWord(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, word string) error {
	r.calls = append(r.calls, "DeleteAllByLivestreamIDMatchingNGWord")
	r.gotLivestreamID = livestreamID
	r.deletedWords = append(r.deletedWords, word)
	return r.deleteErr
}

func (r *fakeLivecommentRepository) Create(ctx context.Context, q repository.Querier, livecomment *model.LivecommentModel) (model.LivecommentID, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = livecomment
	return r.createID, r.createErr
}

func (r *fakeLivecommentRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamID")
	r.gotLivestreamID = livestreamID
	return r.livecomments, r.err
}

func (r *fakeLivecommentRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit model.Limit) ([]*model.Livecomment, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamIDLimited")
	r.gotLivestreamID = livestreamID
	r.gotLimit = limit
	return r.livecomments, r.err
}
