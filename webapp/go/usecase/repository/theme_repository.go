package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type ThemeRepository interface {
	// FindByUserID はテーマが存在しない場合 ErrNotFound を返す。
	FindByUserID(ctx context.Context, q Querier, userID domain.UserID) (*domain.ThemeModel, error)
	// Create はユーザのテーマを登録する。
	Create(ctx context.Context, q Querier, theme *domain.ThemeModel) error
}
