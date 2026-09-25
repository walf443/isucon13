package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type ReactionRepository interface {
	// FindWithDetailsByID はユーザ・ライブ配信を含めたリアクションを返す。
	// リアクションが存在しない場合 ErrNotFound を返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id model.ReactionID) (*model.Reaction, error)
	// FindAllWithDetailsByLivestreamID は指定したライブ配信へのリアクションを、作成日時の降順で返す。
	FindAllWithDetailsByLivestreamID(ctx context.Context, q Querier, livestreamID model.LivestreamID) ([]*model.Reaction, error)
	// FindAllWithDetailsByLivestreamIDLimited は指定したライブ配信へのリアクションを、作成日時の降順で最大 limit 件返す。
	FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q Querier, livestreamID model.LivestreamID, limit int64) ([]*model.Reaction, error)
	// Create はリアクションを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, reaction *model.ReactionModel) (model.ReactionID, error)
	// CountByLivestreamOwnerID は指定したユーザが配信者のライブ配信へのリアクション数を返す。
	CountByLivestreamOwnerID(ctx context.Context, q Querier, userID model.UserID) (int64, error)
	// CountByLivestreamOwnerName は指定した名前のユーザが配信者のライブ配信へのリアクション数を返す。
	CountByLivestreamOwnerName(ctx context.Context, q Querier, name string) (int64, error)
	// FindFavoriteEmojiByLivestreamOwnerName は指定した名前のユーザが配信者のライブ配信で最も多く使われた絵文字を返す。
	// 同数の場合は絵文字名の降順で先頭のものを返す。リアクションが無い場合 ErrNotFound を返す。
	FindFavoriteEmojiByLivestreamOwnerName(ctx context.Context, q Querier, name string) (string, error)
}
