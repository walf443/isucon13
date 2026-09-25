package usecase

import (
	"context"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeLivestreamRepository はテストで設定した関数に処理を委ねる LivestreamRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeLivestreamRepository struct {
	repository.LivestreamRepository

	findByID                    func(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.LivestreamModel, error)
	findAll                     func(ctx context.Context, q repository.Querier) ([]*domain.LivestreamModel, error)
	findAllByUserID             func(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.LivestreamModel, error)
	findAllByIDAndUserID        func(ctx context.Context, q repository.Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.LivestreamModel, error)
	findAllOrderByIDDesc        func(ctx context.Context, q repository.Querier) ([]*domain.LivestreamModel, error)
	findAllOrderByIDDescLimited func(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.LivestreamModel, error)
	create                      func(ctx context.Context, q repository.Querier, livestream *domain.LivestreamModel) (domain.LivestreamID, error)
}

func (r *fakeLivestreamRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.LivestreamModel, error) {
	return r.findByID(ctx, q, id)
}

func (r *fakeLivestreamRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.LivestreamModel, error) {
	return r.findAll(ctx, q)
}

func (r *fakeLivestreamRepository) FindAllByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.LivestreamModel, error) {
	return r.findAllByUserID(ctx, q, userID)
}

func (r *fakeLivestreamRepository) FindAllByIDAndUserID(ctx context.Context, q repository.Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.LivestreamModel, error) {
	return r.findAllByIDAndUserID(ctx, q, id, userID)
}

func (r *fakeLivestreamRepository) FindAllOrderByIDDesc(ctx context.Context, q repository.Querier) ([]*domain.LivestreamModel, error) {
	return r.findAllOrderByIDDesc(ctx, q)
}

func (r *fakeLivestreamRepository) FindAllOrderByIDDescLimited(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.LivestreamModel, error) {
	return r.findAllOrderByIDDescLimited(ctx, q, limit)
}

func (r *fakeLivestreamRepository) Create(ctx context.Context, q repository.Querier, livestream *domain.LivestreamModel) (domain.LivestreamID, error) {
	return r.create(ctx, q, livestream)
}

// newLivestreamRepositoryFindingByID は ID id のライブ配信として livestream (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
// id 以外の ID で呼ばれた場合はテストを失敗させる。
func newLivestreamRepositoryFindingByID(t *testing.T, id domain.LivestreamID, livestream *domain.LivestreamModel, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findByID: func(_ context.Context, _ repository.Querier, gotID domain.LivestreamID) (*domain.LivestreamModel, error) {
			if gotID != id {
				t.Errorf("livestream id = %d, want %d", gotID, id)
			}
			return livestream, err
		},
	}
}
