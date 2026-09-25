package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type UserRepository interface {
	// FindIDByName はユーザが存在しない場合 ErrNotFound を返す。
	FindIDByName(ctx context.Context, q Querier, name string) (domain.UserID, error)
	// FindByID はユーザが存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id domain.UserID) (*domain.UserModel, error)
	// FindByName はユーザが存在しない場合 ErrNotFound を返す。
	FindByName(ctx context.Context, q Querier, name string) (*domain.UserModel, error)
	// Create はユーザを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, user *domain.UserModel) (domain.UserID, error)
	// FindAll は全てのユーザを返す。
	FindAll(ctx context.Context, q Querier) ([]*domain.UserModel, error)

	// FindWithDetailsByID はテーマ・アイコンを含めたユーザを返す。
	// ユーザが存在しない場合 ErrNotFound を返す。テーマが無い場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id domain.UserID) (*domain.User, error)
	// FindWithDetailsByName はテーマ・アイコンを含めたユーザを返す。
	// ユーザが存在しない場合 ErrNotFound を返す。テーマが無い場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByName(ctx context.Context, q Querier, name string) (*domain.User, error)
}
