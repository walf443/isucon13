package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type IconRepository interface {
	// FindAllByUserIDs は userIDs のユーザのアイコンを、アイコンの ID の昇順で返す。アイコンが未登録のユーザの分は結果に含まれない。
	// 同じユーザのアイコンが複数ある場合 (データ不整合) は、全て返す。FindImageByUserID と同じく ID が最小のものを使うこと。
	// userIDs が空の場合はクエリを発行せずに空のスライスを返す。
	FindAllByUserIDs(ctx context.Context, q Querier, userIDs []domain.UserID) ([]*domain.Icon, error)
	// FindImageByUserID はアイコンが登録されていない場合 ErrNotFound を返す。
	FindImageByUserID(ctx context.Context, q Querier, userID domain.UserID) ([]byte, error)
	// Create はアイコンを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, userID domain.UserID, image []byte) (domain.IconID, error)
	DeleteByUserID(ctx context.Context, q Querier, userID domain.UserID) error
}
