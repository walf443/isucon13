package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// livestreamViewersHistoryRow は livestream_viewers_history テーブルの行。
type livestreamViewersHistoryRow struct {
	ID           int64 `gorm:"column:id;primaryKey"`
	UserID       int64 `gorm:"column:user_id"`
	LivestreamID int64 `gorm:"column:livestream_id"`
	// CreatedAt は GORM が登録時刻を自動で設定する名前なので、自動設定を止める (usecase が渡した値をそのまま保存する)
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:false"`
}

// TableName は GORM の複数形の規則 (livestream_viewers_histories) に合わないので明示する。
func (livestreamViewersHistoryRow) TableName() string { return "livestream_viewers_history" }

type livestreamViewersHistoryRepository struct{}

func NewLivestreamViewersHistoryRepository() repository.LivestreamViewersHistoryRepository {
	return &livestreamViewersHistoryRepository{}
}

func (r *livestreamViewersHistoryRepository) Create(ctx context.Context, q repository.Querier, viewer *domain.LivestreamViewersHistory) error {
	row := livestreamViewersHistoryRow{
		UserID:       int64(viewer.UserID),
		LivestreamID: int64(viewer.LivestreamID),
		CreatedAt:    viewer.CreatedAt,
	}
	return dbOf(ctx, q).Create(&row).Error
}

func (r *livestreamViewersHistoryRepository) DeleteByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID domain.UserID, livestreamID domain.LivestreamID) error {
	return dbOf(ctx, q).Where("user_id = ? AND livestream_id = ?", userID, livestreamID).Delete(&livestreamViewersHistoryRow{}).Error
}

func (r *livestreamViewersHistoryRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var cnt int64
	if err := dbOf(ctx, q).Model(&livestreamViewersHistoryRow{}).Where("livestream_id = ?", livestreamID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *livestreamViewersHistoryRepository) CountViewersByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var viewersCount int64
	// ライブ配信統計で使っていた、livestreams と JOIN するクエリのままにしている (ライブ配信が存在しない場合は 0)
	err := dbOf(ctx, q).
		Table("livestreams l").
		Joins("INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id").
		Where("l.id = ?", livestreamID).
		Count(&viewersCount).Error
	if err != nil {
		return 0, err
	}
	return viewersCount, nil
}
