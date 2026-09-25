package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeThemeRepository はテストで設定した関数に処理を委ねる ThemeRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeThemeRepository struct {
	repository.ThemeRepository

	findByUserID func(ctx context.Context, q repository.Querier, userID domain.UserID) (*domain.ThemeModel, error)
	create       func(ctx context.Context, q repository.Querier, theme *domain.ThemeModel) error
}

func (r *fakeThemeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) (*domain.ThemeModel, error) {
	return r.findByUserID(ctx, q, userID)
}

func (r *fakeThemeRepository) Create(ctx context.Context, q repository.Querier, theme *domain.ThemeModel) error {
	return r.create(ctx, q, theme)
}
