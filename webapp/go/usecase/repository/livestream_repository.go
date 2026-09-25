package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type LivestreamRepository interface {
	// FindByID はライブ配信が存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id domain.LivestreamID) (*domain.LivestreamModel, error)
	// FindAllByIDAndUserID は指定した ID かつ指定したユーザが配信者のライブ配信を返す。該当が無ければ空のスライスを返す。
	FindAllByIDAndUserID(ctx context.Context, q Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.LivestreamModel, error)
	// FindAll は全てのライブ配信を返す (順序は不定)。
	FindAll(ctx context.Context, q Querier) ([]*domain.LivestreamModel, error)
	// FindAllByUserID は指定したユーザが配信者のライブ配信を返す (順序は不定)。
	FindAllByUserID(ctx context.Context, q Querier, userID domain.UserID) ([]*domain.LivestreamModel, error)
	// FindWithDetailsByID は配信者・タグを含めたライブ配信を返す。
	// ライブ配信が存在しない場合 ErrNotFound を返す。配信者やタグが欠けている場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id domain.LivestreamID) (*domain.Livestream, error)
	// FindAllWithDetailsByUserID は指定したユーザが配信者のライブ配信を、配信者・タグを含めて返す。
	FindAllWithDetailsByUserID(ctx context.Context, q Querier, userID domain.UserID) ([]*domain.Livestream, error)
	// FindAllWithDetailsByTagIDs は指定したタグのいずれかが付いたライブ配信を、ID の降順で返す。
	// tagIDs は空であってはならない。
	FindAllWithDetailsByTagIDs(ctx context.Context, q Querier, tagIDs []domain.TagID) ([]*domain.Livestream, error)
	// FindAllWithDetails は全てのライブ配信を ID の降順で返す。
	FindAllWithDetails(ctx context.Context, q Querier) ([]*domain.Livestream, error)
	// FindAllWithDetailsLimited は ID の降順で最大 limit 件のライブ配信を返す。
	FindAllWithDetailsLimited(ctx context.Context, q Querier, limit domain.Limit) ([]*domain.Livestream, error)
	// Create はライブ配信を登録し、その ID を返す。
	Create(ctx context.Context, q Querier, livestream *domain.LivestreamModel) (domain.LivestreamID, error)
	// AddTag はライブ配信にタグを付ける。
	AddTag(ctx context.Context, q Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error
}
