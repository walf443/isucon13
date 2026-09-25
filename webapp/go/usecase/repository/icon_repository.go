package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type IconRepository interface {
	// FindImageByUserID はアイコンが登録されていない場合 ErrNotFound を返す。
	FindImageByUserID(ctx context.Context, q Querier, userID model.UserID) ([]byte, error)
	// Create はアイコンを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, userID model.UserID, image []byte) (model.IconID, error)
	DeleteByUserID(ctx context.Context, q Querier, userID model.UserID) error
}
