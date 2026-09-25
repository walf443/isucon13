package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type NGWordRepository interface {
	// FindAllByLivestreamID はライブ配信に登録された NG ワードを (登録したユーザによらず) 全て返す。
	FindAllByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) ([]*domain.NGWordModel, error)
	// FindAllByUserIDAndLivestreamID は指定したユーザがライブ配信に登録した NG ワードを、作成日時の降順で返す。
	FindAllByUserIDAndLivestreamID(ctx context.Context, q Querier, userID domain.UserID, livestreamID domain.LivestreamID) ([]*domain.NGWordModel, error)
	// Matches は comment に NG ワード word が含まれるかを返す。
	// 判定は MySQL の LIKE '%word%' で行う (照合順序に従い大文字小文字を区別しない。word 中の % や _ はワイルドカードになる)。
	Matches(ctx context.Context, q Querier, comment string, word string) (bool, error)
	// Create は NG ワードを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, ngWord *domain.NGWordModel) (domain.NGWordID, error)
}
