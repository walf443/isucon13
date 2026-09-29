package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// UserFiller はユーザにテーマ・アイコンを埋めた domain.UserDetail を組み立てる。
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

// Fill は users にテーマ・アイコンを埋めた domain.UserDetail を、ユーザの ID ごとに返す。
// 同じユーザが複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
// アイコンのハッシュは domain.UserIconHash で決める (未登録の場合は defaultIconHash)。
//
// テーマが無いのはデータ不整合なので、呼び出し側はこのエラーをユーザ不在 (404) として扱わないこと。
func (f *UserFiller) Fill(ctx context.Context, q repository.Querier, users []*domain.User) (map[domain.UserID]*domain.UserDetail, error) {
	userDetails := make(map[domain.UserID]*domain.UserDetail, len(users))
	for _, user := range users {
		if _, ok := userDetails[user.ID]; ok {
			continue
		}

		theme, err := f.themeRepo.FindByUserID(ctx, q, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get theme of user %d: %w", user.ID, asMissingDetail(err))
		}

		image, err := f.iconRepo.FindImageByUserID(ctx, q, user.ID)
		registered := true
		if errors.Is(err, repository.ErrNotFound) {
			registered = false
		} else if err != nil {
			return nil, fmt.Errorf("failed to get icon of user %d: %w", user.ID, err)
		}

		userDetails[user.ID] = &domain.UserDetail{
			ID:          user.ID,
			Name:        user.Name,
			DisplayName: user.DisplayName,
			Description: user.Description,
			Theme:       *theme,
			IconHash:    domain.UserIconHash(image, registered, f.defaultIconHash),
		}
	}
	return userDetails, nil
}
