package usecase

import (
	"context"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeUserRepository はテストで設定した関数に処理を委ねる UserRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeUserRepository struct {
	repository.UserRepository

	findIDByName          func(ctx context.Context, q repository.Querier, name string) (domain.UserID, error)
	findByID              func(ctx context.Context, q repository.Querier, id domain.UserID) (*domain.UserModel, error)
	findByName            func(ctx context.Context, q repository.Querier, name string) (*domain.UserModel, error)
	findAll               func(ctx context.Context, q repository.Querier) ([]*domain.UserModel, error)
	findWithDetailsByID   func(ctx context.Context, q repository.Querier, id domain.UserID) (*domain.User, error)
	findWithDetailsByName func(ctx context.Context, q repository.Querier, name string) (*domain.User, error)
	create                func(ctx context.Context, q repository.Querier, user *domain.UserModel) (domain.UserID, error)
}

func (r *fakeUserRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (domain.UserID, error) {
	return r.findIDByName(ctx, q, name)
}

func (r *fakeUserRepository) FindByID(ctx context.Context, q repository.Querier, id domain.UserID) (*domain.UserModel, error) {
	return r.findByID(ctx, q, id)
}

func (r *fakeUserRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*domain.UserModel, error) {
	return r.findByName(ctx, q, name)
}

func (r *fakeUserRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.UserModel, error) {
	return r.findAll(ctx, q)
}

func (r *fakeUserRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id domain.UserID) (*domain.User, error) {
	return r.findWithDetailsByID(ctx, q, id)
}

func (r *fakeUserRepository) FindWithDetailsByName(ctx context.Context, q repository.Querier, name string) (*domain.User, error) {
	return r.findWithDetailsByName(ctx, q, name)
}

func (r *fakeUserRepository) Create(ctx context.Context, q repository.Querier, user *domain.UserModel) (domain.UserID, error) {
	return r.create(ctx, q, user)
}

// newUserRepositoryFindingID はユーザ名 name から id (失敗させる場合は err) を引く fakeUserRepository を返す。
// name 以外のユーザ名で呼ばれた場合はテストを失敗させる。
func newUserRepositoryFindingID(t *testing.T, name string, id domain.UserID, err error) *fakeUserRepository {
	return &fakeUserRepository{
		findIDByName: func(_ context.Context, _ repository.Querier, gotName string) (domain.UserID, error) {
			if gotName != name {
				t.Errorf("name = %q, want %q", gotName, name)
			}
			return id, err
		},
	}
}
