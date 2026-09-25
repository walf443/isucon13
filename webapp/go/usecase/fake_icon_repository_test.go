package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeIconRepository はテストで設定した関数に処理を委ねる IconRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeIconRepository struct {
	repository.IconRepository

	findImageByUserID func(ctx context.Context, q repository.Querier, userID domain.UserID) ([]byte, error)
	create            func(ctx context.Context, q repository.Querier, userID domain.UserID, image []byte) (domain.IconID, error)
	deleteByUserID    func(ctx context.Context, q repository.Querier, userID domain.UserID) error
}

func (r *fakeIconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]byte, error) {
	return r.findImageByUserID(ctx, q, userID)
}

func (r *fakeIconRepository) Create(ctx context.Context, q repository.Querier, userID domain.UserID, image []byte) (domain.IconID, error) {
	return r.create(ctx, q, userID, image)
}

func (r *fakeIconRepository) DeleteByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) error {
	return r.deleteByUserID(ctx, q, userID)
}
