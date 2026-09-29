package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type TagRepository interface {
	// FindAllByIDs は ids のタグを返す (順序は不定)。存在しない ID は結果に含まれない。
	// ids が空の場合はクエリを発行せずに空のスライスを返す。
	FindAllByIDs(ctx context.Context, q Querier, ids []domain.TagID) ([]*domain.Tag, error)
	FindAll(ctx context.Context, q Querier) ([]*domain.Tag, error)
	// FindIDsByName は指定した名前のタグの ID を返す。該当が無い場合は空のスライスを返す。
	FindIDsByName(ctx context.Context, q Querier, name string) ([]domain.TagID, error)
	// FindByID はタグが存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id domain.TagID) (*domain.Tag, error)
}
