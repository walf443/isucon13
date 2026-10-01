package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// livestreamRow は livestreams テーブルの行。
type livestreamRow struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	UserID       int64  `gorm:"column:user_id"`
	Title        string `gorm:"column:title"`
	Description  string `gorm:"column:description"`
	PlaylistUrl  string `gorm:"column:playlist_url"`
	ThumbnailUrl string `gorm:"column:thumbnail_url"`
	StartAt      int64  `gorm:"column:start_at"`
	EndAt        int64  `gorm:"column:end_at"`
}

func (livestreamRow) TableName() string { return "livestreams" }

func (r *livestreamRow) toDomain() *domain.Livestream {
	return &domain.Livestream{
		ID:           domain.LivestreamID(r.ID),
		UserID:       domain.UserID(r.UserID),
		Title:        r.Title,
		Description:  r.Description,
		PlaylistUrl:  r.PlaylistUrl,
		ThumbnailUrl: r.ThumbnailUrl,
		StartAt:      r.StartAt,
		EndAt:        r.EndAt,
	}
}

type livestreamRepository struct{}

func NewLivestreamRepository() repository.LivestreamRepository {
	return &livestreamRepository{}
}

func (r *livestreamRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.Livestream, error) {
	var row livestreamRow
	err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Where("id = ?", id).Take(&row).Error
	if err != nil {
		return nil, notFound(err)
	}
	return row.toDomain(), nil
}

func (r *livestreamRepository) FindAllByIDAndUserID(ctx context.Context, q repository.Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.Livestream, error) {
	var rows []*livestreamRow
	if err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Where("id = ? AND user_id = ?", id, userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamRow).toDomain), nil
}

func (r *livestreamRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	var rows []*livestreamRow
	if err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamRow).toDomain), nil
}

func (r *livestreamRepository) FindAllByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.Livestream, error) {
	var rows []*livestreamRow
	if err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamRow).toDomain), nil
}

func (r *livestreamRepository) FindAllOrderByIDDesc(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	var rows []*livestreamRow
	if err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamRow).toDomain), nil
}

func (r *livestreamRepository) FindAllOrderByIDDescLimited(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.Livestream, error) {
	var rows []*livestreamRow
	if err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Order("id DESC").Limit(int(limit)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamRow).toDomain), nil
}

func (r *livestreamRepository) Create(ctx context.Context, q repository.Querier, livestream *domain.Livestream) (domain.LivestreamID, error) {
	row := livestreamRow{
		UserID:       int64(livestream.UserID),
		Title:        livestream.Title,
		Description:  livestream.Description,
		PlaylistUrl:  livestream.PlaylistUrl,
		ThumbnailUrl: livestream.ThumbnailUrl,
		StartAt:      livestream.StartAt,
		EndAt:        livestream.EndAt,
	}
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.LivestreamID(row.ID), nil
}

func (r *livestreamRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.LivestreamID) ([]*domain.Livestream, error) {
	rows, err := findIn(ids, func(chunk []domain.LivestreamID) ([]*livestreamRow, error) {
		var rows []*livestreamRow
		err := dbOf(ctx, q).Select("id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at").Where("id IN ?", chunk).Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*livestreamRow).toDomain), nil
}
