package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// reactionRow は reactions テーブルの行。
type reactionRow struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	UserID       int64  `gorm:"column:user_id"`
	LivestreamID int64  `gorm:"column:livestream_id"`
	EmojiName    string `gorm:"column:emoji_name"`
	// CreatedAt は GORM が登録時刻を自動で設定する名前なので、自動設定を止める (usecase が渡した値をそのまま保存する)
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:false"`
}

func (reactionRow) TableName() string { return "reactions" }

func (r *reactionRow) toDomain() *domain.Reaction {
	return &domain.Reaction{
		ID:           domain.ReactionID(r.ID),
		EmojiName:    r.EmojiName,
		UserID:       domain.UserID(r.UserID),
		LivestreamID: domain.LivestreamID(r.LivestreamID),
		CreatedAt:    r.CreatedAt,
	}
}

// favoriteEmojiRow は、最も使われた絵文字を引く集計クエリの結果の行。
type favoriteEmojiRow struct {
	EmojiName string `gorm:"column:emoji_name"`
}

type reactionRepository struct{}

func NewReactionRepository() repository.ReactionRepository {
	return &reactionRepository{}
}

func (r *reactionRepository) FindByID(ctx context.Context, q repository.Querier, id domain.ReactionID) (*domain.Reaction, error) {
	var row reactionRow
	err := dbOf(ctx, q).Select("id, emoji_name, user_id, livestream_id, created_at").Where("id = ?", id).Take(&row).Error
	if err != nil {
		return nil, notFound(err)
	}
	return row.toDomain(), nil
}

func (r *reactionRepository) FindAllByLivestreamIDOrderByCreatedAtDesc(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Reaction, error) {
	var rows []*reactionRow
	err := dbOf(ctx, q).
		Select("id, emoji_name, user_id, livestream_id, created_at").
		Where("livestream_id = ?", livestreamID).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*reactionRow).toDomain), nil
}

func (r *reactionRepository) FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Reaction, error) {
	var rows []*reactionRow
	err := dbOf(ctx, q).
		Select("id, emoji_name, user_id, livestream_id, created_at").
		Where("livestream_id = ?", livestreamID).
		Order("created_at DESC").
		Limit(int(limit)).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*reactionRow).toDomain), nil
}

func (r *reactionRepository) Create(ctx context.Context, q repository.Querier, reaction *domain.Reaction) (domain.ReactionID, error) {
	row := reactionRow{
		UserID:       int64(reaction.UserID),
		LivestreamID: int64(reaction.LivestreamID),
		EmojiName:    reaction.EmojiName,
		CreatedAt:    reaction.CreatedAt,
	}
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.ReactionID(row.ID), nil
}

func (r *reactionRepository) CountByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID domain.UserID) (int64, error) {
	var reactions int64
	err := dbOf(ctx, q).
		Table("users u").
		Joins("INNER JOIN livestreams l ON l.user_id = u.id").
		Joins("INNER JOIN reactions r ON r.livestream_id = l.id").
		Where("u.id = ?", userID).
		Count(&reactions).Error
	if err != nil {
		return 0, err
	}
	return reactions, nil
}

func (r *reactionRepository) CountByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	var reactions int64
	err := dbOf(ctx, q).
		Table("users u").
		Joins("INNER JOIN livestreams l ON l.user_id = u.id").
		Joins("INNER JOIN reactions r ON r.livestream_id = l.id").
		Where("u.name = ?", name).
		Count(&reactions).Error
	if err != nil {
		return 0, err
	}
	return reactions, nil
}

func (r *reactionRepository) FindFavoriteEmojiByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (string, error) {
	var row favoriteEmojiRow
	// 同数の場合は絵文字名の降順で先頭のものを返す (ORDER BY の 2 つ目のキー)
	err := dbOf(ctx, q).
		Table("users u").
		Select("r.emoji_name").
		Joins("INNER JOIN livestreams l ON l.user_id = u.id").
		Joins("INNER JOIN reactions r ON r.livestream_id = l.id").
		Where("u.name = ?", name).
		Group("emoji_name").
		Order("COUNT(*) DESC, emoji_name DESC").
		Take(&row).Error
	if err != nil {
		return "", notFound(err)
	}
	return row.EmojiName, nil
}

func (r *reactionRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var reactions int64
	// CountTotalByLivestreamID とは JOIN 条件の書き方が違う (移行前のクエリの書き方を保っている)
	err := dbOf(ctx, q).
		Table("livestreams l").
		Joins("INNER JOIN reactions r ON l.id = r.livestream_id").
		Where("l.id = ?", livestreamID).
		Count(&reactions).Error
	if err != nil {
		return 0, err
	}
	return reactions, nil
}

func (r *reactionRepository) CountTotalByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	var totalReactions int64
	err := dbOf(ctx, q).
		Table("livestreams l").
		Joins("INNER JOIN reactions r ON r.livestream_id = l.id").
		Where("l.id = ?", livestreamID).
		Count(&totalReactions).Error
	if err != nil {
		return 0, err
	}
	return totalReactions, nil
}
