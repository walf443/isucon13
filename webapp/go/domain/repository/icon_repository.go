package repository

import "context"

type IconRepository interface {
	// FindImageByUserID はアイコンが登録されていない場合 ErrNotFound を返す。
	FindImageByUserID(ctx context.Context, q Querier, userID int64) ([]byte, error)
}
