package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
)

// livecommentRow は livecomments テーブルの行。
type livecommentRow struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	UserID       int64  `gorm:"column:user_id"`
	LivestreamID int64  `gorm:"column:livestream_id"`
	Comment      string `gorm:"column:comment"`
	Tip          int64  `gorm:"column:tip"`
	// CreatedAt は GORM が登録時刻を自動で設定する名前なので、自動設定を止める (usecase が渡した値をそのまま保存する)
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:false"`
}

func (livecommentRow) TableName() string { return "livecomments" }

func (r *livecommentRow) toDomain() *domain.Livecomment {
	return &domain.Livecomment{
		ID:           domain.LivecommentID(r.ID),
		UserID:       domain.UserID(r.UserID),
		LivestreamID: domain.LivestreamID(r.LivestreamID),
		Comment:      r.Comment,
		Tip:          r.Tip,
		CreatedAt:    r.CreatedAt,
	}
}

type livecommentRepository struct{}

func NewLivecommentRepository() repository.LivecommentRepository {
	return &livecommentRepository{}
}

func (r *livecommentRepository) FindAllByLivestreamIDOrderByCreatedAtDesc(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Livecomment, error) {
	var rows []*livecommentRow
	err := dbOf(ctx, q).
		Select("id, user_id, livestream_id, comment, tip, created_at").
		Where("livestream_id = ?", livestreamID).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*livecommentRow).toDomain), nil
}

func (r *livecommentRepository) FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Livecomment, error) {
	var rows []*livecommentRow
	err := dbOf(ctx, q).
		Select("id, user_id, livestream_id, comment, tip, created_at").
		Where("livestream_id = ?", livestreamID).
		Order("created_at DESC").
		Limit(int(limit)).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*livecommentRow).toDomain), nil
}

func (r *livecommentRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivecommentID) (*domain.Livecomment, error) {
	var row livecommentRow
	err := dbOf(ctx, q).Select("id, user_id, livestream_id, comment, tip, created_at").Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *livecommentRepository) Create(ctx context.Context, q repository.Querier, livecomment *domain.Livecomment) (domain.LivecommentID, error) {
	row := livecommentRow{
		UserID:       int64(livecomment.UserID),
		LivestreamID: int64(livecomment.LivestreamID),
		Comment:      livecomment.Comment,
		Tip:          livecomment.Tip,
		CreatedAt:    livecomment.CreatedAt,
	}
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.LivecommentID(row.ID), nil
}

func (r *livecommentRepository) DeleteAllByLivestreamIDMatchingNGWord(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, word string) error {
	// 移行前と同じクエリの流れ (全ライブコメントを取得し、1件ずつ条件付きで DELETE する) を保っている
	var rows []*livecommentRow
	if err := dbOf(ctx, q).Select("id, user_id, livestream_id, comment, tip, created_at").Find(&rows).Error; err != nil {
		return fmt.Errorf("failed to get livecomments: %w", err)
	}

	for _, row := range rows {
		// スパム判定は MySQL の LIKE で行う。照合順序やワイルドカードの扱いまで含めて移行前と同じ判定になるよう、
		// Go には移さずに、SQL をそのまま実行する
		query := `
			DELETE FROM livecomments
			WHERE
			id = ? AND
			livestream_id = ? AND
			(SELECT COUNT(*)
			FROM
			(SELECT ? AS text) AS texts
			INNER JOIN
			(SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
			ON texts.text LIKE patterns.pattern) >= 1;
			`
		if err := dbOf(ctx, q).Exec(query, row.ID, livestreamID, row.Comment, word).Error; err != nil {
			return fmt.Errorf("failed to delete old livecomments that hit spams: %w", err)
		}
	}
	return nil
}

func (r *livecommentRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Livecomment, error) {
	var rows []*livecommentRow
	if err := dbOf(ctx, q).Select("id, user_id, livestream_id, comment, tip, created_at").Where("livestream_id = ?", livestreamID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*livecommentRow).toDomain), nil
}

func (r *livecommentRepository) SumTipByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID domain.UserID) (int64, error) {
	var tips int64
	err := dbOf(ctx, q).
		Table("users u").
		Select("IFNULL(SUM(l2.tip), 0)").
		Joins("INNER JOIN livestreams l ON l.user_id = u.id").
		Joins("INNER JOIN livecomments l2 ON l2.livestream_id = l.id").
		Where("u.id = ?", userID).
		Scan(&tips).Error
	if err != nil {
		return 0, err
	}
	return tips, nil
}

func (r *livecommentRepository) SumTip(ctx context.Context, q repository.Querier) (int64, error) {
	var totalTip int64
	if err := dbOf(ctx, q).Model(&livecommentRow{}).Select("IFNULL(SUM(tip), 0)").Scan(&totalTip).Error; err != nil {
		return 0, err
	}
	return totalTip, nil
}

func (r *livecommentRepository) SumTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var totalTips int64
	err := dbOf(ctx, q).
		Table("livestreams l").
		Select("IFNULL(SUM(l2.tip), 0)").
		Joins("INNER JOIN livecomments l2 ON l.id = l2.livestream_id").
		Where("l.id = ?", livestreamID).
		Scan(&totalTips).Error
	if err != nil {
		return 0, err
	}
	return totalTips, nil
}

func (r *livecommentRepository) MaxTipByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var maxTip int64
	err := dbOf(ctx, q).
		Table("livestreams l").
		Select("IFNULL(MAX(tip), 0)").
		Joins("INNER JOIN livecomments l2 ON l2.livestream_id = l.id").
		Where("l.id = ?", livestreamID).
		Scan(&maxTip).Error
	if err != nil {
		return 0, err
	}
	return maxTip, nil
}

func (r *livecommentRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.LivecommentID) ([]*domain.Livecomment, error) {
	rows, err := findIn(ids, func(chunk []domain.LivecommentID) ([]*livecommentRow, error) {
		var rows []*livecommentRow
		err := dbOf(ctx, q).Select("id, user_id, livestream_id, comment, tip, created_at").Where("id IN ?", chunk).Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*livecommentRow).toDomain), nil
}
