package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type IconRepository interface {
	// FindAllByUserIDs は userIDs のユーザのアイコンを返す (順序は不定)。アイコンが未登録のユーザの分は結果に含まれない。
	// userIDs が空の場合はクエリを発行せずに空のスライスを返す。
	FindAllByUserIDs(ctx context.Context, q Querier, userIDs []domain.UserID) ([]*domain.Icon, error)
	// FindImageByUserID はアイコンが登録されていない場合 ErrNotFound を返す。
	FindImageByUserID(ctx context.Context, q Querier, userID domain.UserID) ([]byte, error)
	// Create はアイコンを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, userID domain.UserID, image []byte) (domain.IconID, error)
	DeleteByUserID(ctx context.Context, q Querier, userID domain.UserID) error
}
