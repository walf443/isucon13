package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type LivestreamRepository interface {
	// FindWithDetailsByID は配信者・タグを含めたライブ配信を返す。
	// ライブ配信が存在しない場合 ErrNotFound を返す。配信者やタグが欠けている場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id int64) (*model.Livestream, error)
	// FindAllWithDetailsByUserID は指定したユーザが配信者のライブ配信を、配信者・タグを含めて返す。
	FindAllWithDetailsByUserID(ctx context.Context, q Querier, userID int64) ([]*model.Livestream, error)
	// FindAllWithDetailsByTagIDs は指定したタグのいずれかが付いたライブ配信を、ID の降順で返す。
	// tagIDs は空であってはならない。
	FindAllWithDetailsByTagIDs(ctx context.Context, q Querier, tagIDs []int64) ([]*model.Livestream, error)
	// FindAllWithDetails は全てのライブ配信を ID の降順で返す。
	FindAllWithDetails(ctx context.Context, q Querier) ([]*model.Livestream, error)
	// FindAllWithDetailsLimited は ID の降順で最大 limit 件のライブ配信を返す。
	FindAllWithDetailsLimited(ctx context.Context, q Querier, limit int64) ([]*model.Livestream, error)
}
