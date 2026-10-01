package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type UserRepository interface {
	// FindAllByIDs は ids のユーザを返す (順序は不定)。存在しない ID は結果に含まれない。
	// ids が空の場合はクエリを発行せずに空のスライスを返す。
	FindAllByIDs(ctx context.Context, q Querier, ids []domain.UserID) ([]*domain.User, error)
	// FindIDByName はユーザが存在しない場合 ErrNotFound を返す。
	FindIDByName(ctx context.Context, q Querier, name domain.Username) (domain.UserID, error)
	// FindByID はユーザが存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id domain.UserID) (*domain.User, error)
	// FindByName はユーザが存在しない場合 ErrNotFound を返す。
	FindByName(ctx context.Context, q Querier, name domain.Username) (*domain.User, error)
	// Create はユーザを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, user *domain.User) (domain.UserID, error)
	// FindAll は全てのユーザを返す。
	FindAll(ctx context.Context, q Querier) ([]*domain.User, error)
}
