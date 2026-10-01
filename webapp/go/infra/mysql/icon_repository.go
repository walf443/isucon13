package mysql

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
)

// iconRow は icons テーブルの行。
type iconRow struct {
	ID     int64  `gorm:"column:id;primaryKey"`
	UserID int64  `gorm:"column:user_id"`
	Image  []byte `gorm:"column:image"`
}

func (iconRow) TableName() string { return "icons" }

func (r *iconRow) toDomain() *domain.Icon {
	return &domain.Icon{ID: domain.IconID(r.ID), UserID: domain.UserID(r.UserID), Image: r.Image}
}

type iconRepository struct{}

func NewIconRepository() repository.IconRepository {
	return &iconRepository{}
}

func (r *iconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]byte, error) {
	var row iconRow
	// 同じユーザのアイコンが複数ある場合 (データ不整合) は、ID が最小の行を使う (First は ID の昇順の先頭)
	err := dbOf(ctx, q).Select("image").Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.Image, nil
}

func (r *iconRepository) Create(ctx context.Context, q repository.Querier, userID domain.UserID, image []byte) (domain.IconID, error) {
	row := iconRow{UserID: int64(userID), Image: image}
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.IconID(row.ID), nil
}

func (r *iconRepository) DeleteByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) error {
	return dbOf(ctx, q).Where("user_id = ?", userID).Delete(&iconRow{}).Error
}

func (r *iconRepository) FindAllByUserIDs(ctx context.Context, q repository.Querier, userIDs []domain.UserID) ([]*domain.Icon, error) {
	rows, err := findIn(userIDs, func(chunk []domain.UserID) ([]*iconRow, error) {
		var rows []*iconRow
		err := dbOf(ctx, q).Select("id, user_id, image").Where("user_id IN ?", chunk).Order("id").Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	// ID を分割して引いた場合に、結果全体で ID の昇順になるようにする
	slices.SortFunc(rows, func(a, b *iconRow) int { return cmp.Compare(a.ID, b.ID) })
	return mapRows(rows, (*iconRow).toDomain), nil
}
