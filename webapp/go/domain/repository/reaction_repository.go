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
}
