package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// themeRow は themes テーブルの行。
type themeRow struct {
	ID       int64 `gorm:"column:id;primaryKey"`
	UserID   int64 `gorm:"column:user_id"`
	DarkMode bool  `gorm:"column:dark_mode"`
}

func (themeRow) TableName() string { return "themes" }

func (r *themeRow) toDomain() *domain.Theme {
	return &domain.Theme{ID: domain.ThemeID(r.ID), UserID: domain.UserID(r.UserID), DarkMode: r.DarkMode}
}

type themeRepository struct{}

func NewThemeRepository() repository.ThemeRepository {
	return &themeRepository{}
}

func (r *themeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) (*domain.Theme, error) {
	var row themeRow
	err := dbOf(ctx, q).Select("id, user_id, dark_mode").Where("user_id = ?", userID).Take(&row).Error
	if err != nil {
		return nil, notFound(err)
	}
	return row.toDomain(), nil
}

func (r *themeRepository) Create(ctx context.Context, q repository.Querier, theme *domain.Theme) error {
	row := themeRow{UserID: int64(theme.UserID), DarkMode: theme.DarkMode}
	return dbOf(ctx, q).Create(&row).Error
}

func (r *themeRepository) FindAllByUserIDs(ctx context.Context, q repository.Querier, userIDs []domain.UserID) ([]*domain.Theme, error) {
	rows, err := findIn(userIDs, func(chunk []domain.UserID) ([]*themeRow, error) {
		var rows []*themeRow
		err := dbOf(ctx, q).Select("id, user_id, dark_mode").Where("user_id IN ?", chunk).Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*themeRow).toDomain), nil
}
