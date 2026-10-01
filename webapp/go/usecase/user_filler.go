package usecase

import (
	"context"
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
// テーマ・アイコンはそれぞれ 1 回のクエリでまとめて取得する。並び順は呼び出し側で決める。
// アイコンのハッシュは domain.UserIconHash で決める (未登録の場合は defaultIconHash)。
//
// テーマが無いのはデータ不整合なので errMissingDetail を返す (呼び出し側はユーザ不在 (404) として扱わないこと)。
func (f *UserFiller) Fill(ctx context.Context, q repository.Querier, users []*domain.User) (map[domain.UserID]*domain.UserDetail, error) {
	userIDs := uniqueKeys(users, func(user *domain.User) domain.UserID { return user.ID })

	themes, err := f.themeRepo.FindAllByUserIDs(ctx, q, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get themes: %w", err)
	}
	themesByUserID := indexBy(themes, func(theme *domain.Theme) domain.UserID { return theme.UserID })

	icons, err := f.iconRepo.FindAllByUserIDs(ctx, q, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get icons: %w", err)
	}
	// 同じユーザのアイコンが複数ある場合 (データ不整合) は、FindImageByUserID と同じく ID が最小の (先頭の) ものを使う
	iconsByUserID := make(map[domain.UserID]*domain.Icon, len(icons))
	for _, icon := range icons {
		if _, ok := iconsByUserID[icon.UserID]; !ok {
			iconsByUserID[icon.UserID] = icon
		}
	}

	userDetails := make(map[domain.UserID]*domain.UserDetail, len(userIDs))
	for _, user := range users {
		if _, ok := userDetails[user.ID]; ok {
			continue
		}

		theme, ok := themesByUserID[user.ID]
		if !ok {
			return nil, fmt.Errorf("failed to get theme of user %d: %w", user.ID, missingDetail())
		}

		var image []byte
		icon, registered := iconsByUserID[user.ID]
		if registered {
			image = icon.Image
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
