package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type ThemeRepository interface {
	// FindByUserID はテーマが存在しない場合 ErrNotFound を返す。
	FindByUserID(ctx context.Context, q Querier, userID int64) (*model.ThemeModel, error)
}
