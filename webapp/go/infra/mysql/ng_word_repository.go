package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// ngWordRow は ng_words テーブルの行。
type ngWordRow struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	UserID       int64  `gorm:"column:user_id"`
	LivestreamID int64  `gorm:"column:livestream_id"`
	Word         string `gorm:"column:word"`
	// CreatedAt は GORM が登録時刻を自動で設定する名前なので、自動設定を止める (usecase が渡した値をそのまま保存する)
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:false"`
}

func (ngWordRow) TableName() string { return "ng_words" }

func (r *ngWordRow) toDomain() *domain.NGWord {
	return &domain.NGWord{
		ID:           domain.NGWordID(r.ID),
		UserID:       domain.UserID(r.UserID),
		LivestreamID: domain.LivestreamID(r.LivestreamID),
		Word:         r.Word,
		CreatedAt:    r.CreatedAt,
	}
}

type ngWordRepository struct{}

func NewNGWordRepository() repository.NGWordRepository {
	return &ngWordRepository{}
}

func (r *ngWordRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.NGWord, error) {
	var rows []*ngWordRow
	if err := dbOf(ctx, q).Select("id, user_id, livestream_id, word, created_at").Where("livestream_id = ?", livestreamID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*ngWordRow).toDomain), nil
}

func (r *ngWordRepository) FindAllByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID domain.UserID, livestreamID domain.LivestreamID) ([]*domain.NGWord, error) {
	var rows []*ngWordRow
	err := dbOf(ctx, q).
		Select("id, user_id, livestream_id, word, created_at").
		Where("user_id = ? AND livestream_id = ?", userID, livestreamID).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*ngWordRow).toDomain), nil
}

func (r *ngWordRepository) Matches(ctx context.Context, q repository.Querier, comment string, word string) (bool, error) {
	// スパム判定は MySQL の LIKE で行う。照合順序 (大文字小文字を区別しない) やワイルドカードの扱いまで含めて
	// 移行前と同じ判定になるよう、Go には移さずに、SQL をそのまま実行する
	query := `
		SELECT COUNT(*)
		FROM
		(SELECT ? AS text) AS texts
		INNER JOIN
		(SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
		ON texts.text LIKE patterns.pattern;
		`
	var hitSpam int
	if err := dbOf(ctx, q).Raw(query, comment, word).Scan(&hitSpam).Error; err != nil {
		return false, err
	}
	return hitSpam >= 1, nil
}

func (r *ngWordRepository) Create(ctx context.Context, q repository.Querier, ngWord *domain.NGWord) (domain.NGWordID, error) {
	row := ngWordRow{
		UserID:       int64(ngWord.UserID),
		LivestreamID: int64(ngWord.LivestreamID),
		Word:         ngWord.Word,
		CreatedAt:    ngWord.CreatedAt,
	}
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.NGWordID(row.ID), nil
}
