package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type UserRepository interface {
	// FindIDByName はユーザが存在しない場合 ErrNotFound を返す。
	FindIDByName(ctx context.Context, q Querier, name string) (int64, error)
	// FindByName はユーザが存在しない場合 ErrNotFound を返す。
	FindByName(ctx context.Context, q Querier, name string) (*model.UserModel, error)
}
