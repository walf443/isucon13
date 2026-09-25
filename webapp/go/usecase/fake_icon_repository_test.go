package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeIconRepository struct {
	image     []byte
	err       error
	createID  model.IconID
	createErr error
	deleteErr error

	// calls は呼ばれたメソッド名を順に記録する
	calls     []string
	gotUserID model.UserID
	gotImage  []byte
}

func (r *fakeIconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID model.UserID) ([]byte, error) {
	r.calls = append(r.calls, "FindImageByUserID")
	r.gotUserID = userID
	return r.image, r.err
}

func (r *fakeIconRepository) Create(ctx context.Context, q repository.Querier, userID model.UserID, image []byte) (model.IconID, error) {
	r.calls = append(r.calls, "Create")
	r.gotUserID = userID
	r.gotImage = image
	return r.createID, r.createErr
}

func (r *fakeIconRepository) DeleteByUserID(ctx context.Context, q repository.Querier, userID model.UserID) error {
	r.calls = append(r.calls, "DeleteByUserID")
	r.gotUserID = userID
	return r.deleteErr
}
