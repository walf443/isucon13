package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type LivecommentRepository interface {
	// FindAllWithDetailsByLivestreamID は指定したライブ配信へのライブコメントを、作成日時の降順で返す。
	FindAllWithDetailsByLivestreamID(ctx context.Context, q Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error)
	// FindAllWithDetailsByLivestreamIDLimited は指定したライブ配信へのライブコメントを、作成日時の降順で最大 limit 件返す。
	FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q Querier, livestreamID model.LivestreamID, limit int64) ([]*model.Livecomment, error)
}
