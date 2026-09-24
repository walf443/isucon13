package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type TagRepository interface {
	FindAll(ctx context.Context, q Querier) ([]*model.TagModel, error)
	// FindIDsByName は指定した名前のタグの ID を返す。該当が無い場合は空のスライスを返す。
	FindIDsByName(ctx context.Context, q Querier, name string) ([]int64, error)
}
