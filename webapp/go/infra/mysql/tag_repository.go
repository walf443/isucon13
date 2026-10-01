package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// tagRow は tags テーブルの行。
type tagRow struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

func (tagRow) TableName() string { return "tags" }

func (r *tagRow) toDomain() *domain.Tag {
	return &domain.Tag{ID: domain.TagID(r.ID), Name: r.Name}
}

type tagRepository struct{}

func NewTagRepository() repository.TagRepository {
	return &tagRepository{}
}

func (r *tagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.Tag, error) {
	var rows []*tagRow
	if err := dbOf(ctx, q).Select("id, name").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*tagRow).toDomain), nil
}

func (r *tagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error) {
	var ids []int64
	if err := dbOf(ctx, q).Model(&tagRow{}).Where("name = ?", name).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	tagIDs := make([]domain.TagID, len(ids))
	for i, id := range ids {
		tagIDs[i] = domain.TagID(id)
	}
	return tagIDs, nil
}

func (r *tagRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.TagID) ([]*domain.Tag, error) {
	rows, err := findIn(ids, func(chunk []domain.TagID) ([]*tagRow, error) {
		var rows []*tagRow
		err := dbOf(ctx, q).Select("id, name").Where("id IN ?", chunk).Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*tagRow).toDomain), nil
}
