package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fillUsers は userModels にテーマ・アイコンを埋めた model.User を、同じ順序で返す。
// アイコンのハッシュは model.UserIconHash で決める (未登録の場合は defaultIconHash)。
//
// テーマが無いのはデータ不整合なので、repository.ErrNotFound (ユーザ不在) には変換しない。
func fillUsers(ctx context.Context, q repository.Querier, userModels []*model.UserModel, defaultIconHash model.IconHash) ([]*model.User, error) {
	users := make([]*model.User, len(userModels))
	for i, userModel := range userModels {
		var theme model.ThemeModel
		if err := q.GetContext(ctx, &theme, "SELECT id, user_id, dark_mode FROM themes WHERE user_id = ?", userModel.ID); err != nil {
			return nil, fmt.Errorf("failed to get theme of user %d: %w", userModel.ID, err)
		}

		var image []byte
		registered := true
		err := q.GetContext(ctx, &image, "SELECT image FROM icons WHERE user_id = ?", userModel.ID)
		if errors.Is(err, sql.ErrNoRows) {
			registered = false
		} else if err != nil {
			return nil, fmt.Errorf("failed to get icon of user %d: %w", userModel.ID, err)
		}

		users[i] = &model.User{
			ID:          userModel.ID,
			Name:        userModel.Name,
			DisplayName: userModel.DisplayName,
			Description: userModel.Description,
			Theme:       theme,
			IconHash:    model.UserIconHash(image, registered, defaultIconHash),
		}
	}
	return users, nil
}
