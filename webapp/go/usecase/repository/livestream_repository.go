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
	// FindAllOrderByIDDesc は全てのライブ配信を ID の降順で返す。
	FindAllOrderByIDDesc(ctx context.Context, q Querier) ([]*domain.LivestreamModel, error)
	// FindAllOrderByIDDescLimited は ID の降順で最大 limit 件のライブ配信を返す。
	FindAllOrderByIDDescLimited(ctx context.Context, q Querier, limit domain.Limit) ([]*domain.LivestreamModel, error)
	// Create はライブ配信を登録し、その ID を返す。
	Create(ctx context.Context, q Querier, livestream *domain.LivestreamModel) (domain.LivestreamID, error)
}
