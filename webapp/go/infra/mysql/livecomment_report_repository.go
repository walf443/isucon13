package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// livecommentReportRow は livecomment_reports テーブルの行。
type livecommentReportRow struct {
	ID            int64 `gorm:"column:id;primaryKey"`
	UserID        int64 `gorm:"column:user_id"`
	LivestreamID  int64 `gorm:"column:livestream_id"`
	LivecommentID int64 `gorm:"column:livecomment_id"`
	// CreatedAt は GORM が登録時刻を自動で設定する名前なので、自動設定を止める (usecase が渡した値をそのまま保存する)
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:false"`
}

func (livecommentReportRow) TableName() string { return "livecomment_reports" }

func (r *livecommentReportRow) toDomain() *domain.LivecommentReport {
	return &domain.LivecommentReport{
		ID:            domain.LivecommentReportID(r.ID),
		UserID:        domain.UserID(r.UserID),
		LivestreamID:  domain.LivestreamID(r.LivestreamID),
		LivecommentID: domain.LivecommentID(r.LivecommentID),
		CreatedAt:     r.CreatedAt,
	}
}

type livecommentReportRepository struct{}

func NewLivecommentReportRepository() repository.LivecommentReportRepository {
	return &livecommentReportRepository{}
}

func (r *livecommentReportRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReport, error) {
	var row livecommentReportRow
	err := dbOf(ctx, q).Select("id, user_id, livestream_id, livecomment_id, created_at").Where("id = ?", id).Take(&row).Error
	if err != nil {
		return nil, notFound(err)
	}
	return row.toDomain(), nil
}

func (r *livecommentReportRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReport, error) {
	var rows []*livecommentReportRow
	if err := dbOf(ctx, q).Select("id, user_id, livestream_id, livecomment_id, created_at").Where("livestream_id = ?", livestreamID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livecommentReportRow).toDomain), nil
}

func (r *livecommentReportRepository) Create(ctx context.Context, q repository.Querier, report *domain.LivecommentReport) (domain.LivecommentReportID, error) {
	row := livecommentReportRow{
		UserID:        int64(report.UserID),
		LivestreamID:  int64(report.LivestreamID),
		LivecommentID: int64(report.LivecommentID),
		CreatedAt:     report.CreatedAt,
	}
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.LivecommentReportID(row.ID), nil
}

func (r *livecommentReportRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var totalReports int64
	err := dbOf(ctx, q).
		Table("livestreams l").
		Joins("INNER JOIN livecomment_reports r ON r.livestream_id = l.id").
		Where("l.id = ?", livestreamID).
		Count(&totalReports).Error
	if err != nil {
		return 0, err
	}
	return totalReports, nil
}
