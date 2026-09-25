package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type NGWordRepository interface {
	// FindAllByUserIDAndLivestreamID は指定したユーザがライブ配信に登録した NG ワードを、作成日時の降順で返す。
	FindAllByUserIDAndLivestreamID(ctx context.Context, q Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error)
}
