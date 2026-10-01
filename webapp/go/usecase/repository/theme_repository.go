package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type ThemeRepository interface {
	// FindAllByUserIDs は userIDs のユーザのテーマを返す (順序は不定)。テーマが無いユーザの分は結果に含まれない。
	// userIDs が空の場合はクエリを発行せずに空のスライスを返す。
	FindAllByUserIDs(ctx context.Context, q Querier, userIDs []domain.UserID) ([]*domain.Theme, error)
	// FindByUserID はテーマが存在しない場合 ErrNotFound を返す。
	FindByUserID(ctx context.Context, q Querier, userID domain.UserID) (*domain.Theme, error)
	// Create はユーザのテーマを登録する。
	Create(ctx context.Context, q Querier, theme *domain.Theme) error
}
