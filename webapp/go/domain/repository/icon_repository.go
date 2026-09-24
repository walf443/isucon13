package repository

import "context"

type IconRepository interface {
	// FindImageByUserID はアイコンが登録されていない場合 ErrNotFound を返す。
	FindImageByUserID(ctx context.Context, q Querier, userID int64) ([]byte, error)
	// Create はアイコンを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, userID int64, image []byte) (int64, error)
	DeleteByUserID(ctx context.Context, q Querier, userID int64) error
}
