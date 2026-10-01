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

	findAll       func(ctx context.Context, q repository.Querier) ([]*domain.Tag, error)
	findIDsByName func(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error)
	findAllByIDs  func(ctx context.Context, q repository.Querier, ids []domain.TagID) ([]*domain.Tag, error)
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.Tag, error) {
	return r.findAll(ctx, q)
}

func (r *fakeTagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error) {
	return r.findIDsByName(ctx, q, name)
}

func (r *fakeTagRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.TagID) ([]*domain.Tag, error) {
	return r.findAllByIDs(ctx, q, ids)
}
