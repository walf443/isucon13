package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeTagRepository はテストで設定した関数に処理を委ねる TagRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeTagRepository struct {
	repository.TagRepository

	findAll       func(ctx context.Context, q repository.Querier) ([]*domain.TagModel, error)
	findIDsByName func(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error)
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.TagModel, error) {
	return r.findAll(ctx, q)
}

func (r *fakeTagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error) {
	return r.findIDsByName(ctx, q, name)
}
