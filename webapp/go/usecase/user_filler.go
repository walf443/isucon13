package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// UserFiller はユーザにテーマ・アイコンを埋めた domain.User を組み立てる。
// 複数の usecase で使うので、1 つ作って共有する。
type UserFiller struct {
	themeRepo repository.ThemeRepository
	iconRepo  repository.IconRepository
	// defaultIconHash はアイコン未登録のユーザに使う既定のアイコンのハッシュ。
	defaultIconHash domain.IconHash
}

func NewUserFiller(themeRepo repository.ThemeRepository, iconRepo repository.IconRepository, defaultIconHash domain.IconHash) *UserFiller {
	return &UserFiller{themeRepo: themeRepo, iconRepo: iconRepo, defaultIconHash: defaultIconHash}
}

// Fill は userModels にテーマ・アイコンを埋めた domain.User を、ユーザの ID ごとに返す。
// 同じユーザが複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
// アイコンのハッシュは domain.UserIconHash で決める (未登録の場合は defaultIconHash)。
//
// テーマが無いのはデータ不整合なので、呼び出し側はこのエラーをユーザ不在 (404) として扱わないこと。
func (f *UserFiller) Fill(ctx context.Context, q repository.Querier, userModels []*domain.User) (map[domain.UserID]*domain.UserDetail, error) {
	users := make(map[domain.UserID]*domain.UserDetail, len(userModels))
	for _, userModel := range userModels {
		if _, ok := users[userModel.ID]; ok {
			continue
		}

		theme, err := f.themeRepo.FindByUserID(ctx, q, userModel.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get theme of user %d: %w", userModel.ID, err)
		}

		image, err := f.iconRepo.FindImageByUserID(ctx, q, userModel.ID)
		registered := true
		if errors.Is(err, repository.ErrNotFound) {
			registered = false
		} else if err != nil {
			return nil, fmt.Errorf("failed to get icon of user %d: %w", userModel.ID, err)
		}

		users[userModel.ID] = &domain.UserDetail{
			ID:          userModel.ID,
			Name:        userModel.Name,
			DisplayName: userModel.DisplayName,
			Description: userModel.Description,
			Theme:       *theme,
			IconHash:    domain.UserIconHash(image, registered, f.defaultIconHash),
		}
	}
	return users, nil
}
