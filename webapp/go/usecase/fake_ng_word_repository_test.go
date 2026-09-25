package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeNGWordRepository はテストで設定した関数に処理を委ねる NGWordRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeNGWordRepository struct {
	repository.NGWordRepository

	findAllByLivestreamID          func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.NGWordModel, error)
	findAllByUserIDAndLivestreamID func(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error)
	matches                        func(ctx context.Context, q repository.Querier, comment string, word string) (bool, error)
	create                         func(ctx context.Context, q repository.Querier, ngWord *model.NGWordModel) (model.NGWordID, error)
}

func (r *fakeNGWordRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	return r.findAllByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeNGWordRepository) FindAllByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	return r.findAllByUserIDAndLivestreamID(ctx, q, userID, livestreamID)
}

func (r *fakeNGWordRepository) Matches(ctx context.Context, q repository.Querier, comment string, word string) (bool, error) {
	return r.matches(ctx, q, comment, word)
}

func (r *fakeNGWordRepository) Create(ctx context.Context, q repository.Querier, ngWord *model.NGWordModel) (model.NGWordID, error) {
	return r.create(ctx, q, ngWord)
}
