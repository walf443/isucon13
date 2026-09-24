package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type TagRepository interface {
	FindAll(ctx context.Context, q Querier) ([]*model.TagModel, error)
}
