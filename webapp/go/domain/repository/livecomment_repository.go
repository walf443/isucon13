package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type LivecommentRepository interface {
	// FindWithDetailsByID はユーザ・ライブ配信を含めたライブコメントを返す。
	// ライブコメントが存在しない場合 ErrNotFound を返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id model.LivecommentID) (*model.Livecomment, error)
	// FindAllWithDetailsByLivestreamID は指定したライブ配信へのライブコメントを、作成日時の降順で返す。
	FindAllWithDetailsByLivestreamID(ctx context.Context, q Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error)
	// FindAllWithDetailsByLivestreamIDLimited は指定したライブ配信へのライブコメントを、作成日時の降順で最大 limit 件返す。
	FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q Querier, livestreamID model.LivestreamID, limit int64) ([]*model.Livecomment, error)
	// Create はライブコメントを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, livecomment *model.LivecommentModel) (model.LivecommentID, error)
}
