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

	findByID                   func(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.LivestreamModel, error)
	findAll                    func(ctx context.Context, q repository.Querier) ([]*domain.LivestreamModel, error)
	findAllByUserID            func(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.LivestreamModel, error)
	findAllByIDAndUserID       func(ctx context.Context, q repository.Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.LivestreamModel, error)
	findWithDetailsByID        func(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.Livestream, error)
	findAllWithDetailsByUserID func(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.Livestream, error)
	findAllWithDetailsByTagIDs func(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.Livestream, error)
	findAllWithDetails         func(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error)
	findAllWithDetailsLimited  func(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.Livestream, error)
	create                     func(ctx context.Context, q repository.Querier, livestream *domain.LivestreamModel) (domain.LivestreamID, error)
	addTag                     func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error
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

func (r *fakeLivestreamRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.Livestream, error) {
	return r.findWithDetailsByID(ctx, q, id)
}

func (r *fakeLivestreamRepository) FindAllWithDetailsByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.Livestream, error) {
	return r.findAllWithDetailsByUserID(ctx, q, userID)
}

func (r *fakeLivestreamRepository) FindAllWithDetailsByTagIDs(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.Livestream, error) {
	return r.findAllWithDetailsByTagIDs(ctx, q, tagIDs)
}

func (r *fakeLivestreamRepository) FindAllWithDetails(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	return r.findAllWithDetails(ctx, q)
}

func (r *fakeLivestreamRepository) FindAllWithDetailsLimited(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.Livestream, error) {
	return r.findAllWithDetailsLimited(ctx, q, limit)
}

func (r *fakeLivestreamRepository) Create(ctx context.Context, q repository.Querier, livestream *domain.LivestreamModel) (domain.LivestreamID, error) {
	return r.create(ctx, q, livestream)
}

func (r *fakeLivestreamRepository) AddTag(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error {
	return r.addTag(ctx, q, livestreamID, tagID)
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
