package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type NGWordRepository interface {
	// FindAllByUserIDAndLivestreamID は指定したユーザがライブ配信に登録した NG ワードを、作成日時の降順で返す。
	FindAllByUserIDAndLivestreamID(ctx context.Context, q Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error)
	// Matches は comment に NG ワード word が含まれるかを返す。
	// 判定は MySQL の LIKE '%word%' で行う (照合順序に従い大文字小文字を区別しない。word 中の % や _ はワイルドカードになる)。
	Matches(ctx context.Context, q Querier, comment string, word string) (bool, error)
}
