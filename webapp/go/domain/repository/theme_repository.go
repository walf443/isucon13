package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type ThemeRepository interface {
	// FindByUserID はテーマが存在しない場合 ErrNotFound を返す。
	FindByUserID(ctx context.Context, q Querier, userID model.UserID) (*model.ThemeModel, error)
	// Create はユーザのテーマを登録する。
	Create(ctx context.Context, q Querier, theme *model.ThemeModel) error
}
