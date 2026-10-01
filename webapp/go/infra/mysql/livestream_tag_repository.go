package mysql

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// livestreamTagRow は livestream_tags テーブルの行。
type livestreamTagRow struct {
	ID           int64 `gorm:"column:id;primaryKey"`
	LivestreamID int64 `gorm:"column:livestream_id"`
	TagID        int64 `gorm:"column:tag_id"`
}

func (livestreamTagRow) TableName() string { return "livestream_tags" }

func (r *livestreamTagRow) toDomain() *domain.LivestreamTag {
	return &domain.LivestreamTag{
		ID:           domain.LivestreamTagID(r.ID),
		LivestreamID: domain.LivestreamID(r.LivestreamID),
		TagID:        domain.TagID(r.TagID),
	}
}

type livestreamTagRepository struct{}

func NewLivestreamTagRepository() repository.LivestreamTagRepository {
	return &livestreamTagRepository{}
}

func (r *livestreamTagRepository) Create(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error {
	row := livestreamTagRow{LivestreamID: int64(livestreamID), TagID: int64(tagID)}
	return dbOf(ctx, q).Create(&row).Error
}

func (r *livestreamTagRepository) FindAllByTagIDs(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error) {
	// 移行前と同じく、空の場合はエラーにする (GORM の IN ? は、空だとエラーにならずに何も返さない)
	if len(tagIDs) == 0 {
		return nil, errors.New("failed to construct IN query: empty slice passed to 'in' query")
	}
	// タグの ID はタグ名で引いたものなので少なく、分割すると結果全体の並び順 (ライブ配信の ID の降順) が崩れるので、分割しない
	var rows []*livestreamTagRow
	if err := dbOf(ctx, q).Select("id, livestream_id, tag_id").Where("tag_id IN ?", tagIDs).Order("livestream_id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamTagRow).toDomain), nil
}

func (r *livestreamTagRepository) FindAllByLivestreamIDs(ctx context.Context, q repository.Querier, livestreamIDs []domain.LivestreamID) ([]*domain.LivestreamTag, error) {
	rows, err := findIn(livestreamIDs, func(chunk []domain.LivestreamID) ([]*livestreamTagRow, error) {
		var rows []*livestreamTagRow
		err := dbOf(ctx, q).Select("id, livestream_id, tag_id").Where("livestream_id IN ?", chunk).Order("id").Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	// ID を分割して引いた場合に、結果全体で ID の昇順になるようにする
	slices.SortFunc(rows, func(a, b *livestreamTagRow) int { return cmp.Compare(a.ID, b.ID) })
	return mapRows(rows, (*livestreamTagRow).toDomain), nil
}
