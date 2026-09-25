package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type UserRepository interface {
	// FindIDByName はユーザが存在しない場合 ErrNotFound を返す。
	FindIDByName(ctx context.Context, q Querier, name string) (model.UserID, error)
	// FindByID はユーザが存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id model.UserID) (*model.UserModel, error)
	// FindByName はユーザが存在しない場合 ErrNotFound を返す。
	FindByName(ctx context.Context, q Querier, name string) (*model.UserModel, error)
	// FindAll は全てのユーザを返す。
	FindAll(ctx context.Context, q Querier) ([]*model.UserModel, error)

	// FindWithDetailsByID はテーマ・アイコンを含めたユーザを返す。
	// ユーザが存在しない場合 ErrNotFound を返す。テーマが無い場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id model.UserID) (*model.User, error)
	// FindWithDetailsByName はテーマ・アイコンを含めたユーザを返す。
	// ユーザが存在しない場合 ErrNotFound を返す。テーマが無い場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByName(ctx context.Context, q Querier, name string) (*model.User, error)
}
