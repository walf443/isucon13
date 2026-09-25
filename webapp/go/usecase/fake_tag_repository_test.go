package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// fakeTagRepository はテストで設定した関数に処理を委ねる TagRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeTagRepository struct {
	repository.TagRepository

	findAll       func(ctx context.Context, q repository.Querier) ([]*model.TagModel, error)
	findIDsByName func(ctx context.Context, q repository.Querier, name string) ([]model.TagID, error)
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.TagModel, error) {
	return r.findAll(ctx, q)
}

func (r *fakeTagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]model.TagID, error) {
	return r.findIDsByName(ctx, q, name)
}
