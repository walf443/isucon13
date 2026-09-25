package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type StatisticsUsecase interface {
	// FindUserStatistics は指定したユーザの配信者としての統計情報を返す。
	// ユーザが存在しない場合 ErrUserNotFound を返す。
	FindUserStatistics(ctx context.Context, username string) (*model.UserStatistics, error)
}

type statisticsUsecase struct {
	txManager       repository.TxManager
	userRepo        repository.UserRepository
	livestreamRepo  repository.LivestreamRepository
	livecommentRepo repository.LivecommentRepository
	reactionRepo    repository.ReactionRepository
	viewerRepo      repository.LivestreamViewersHistoryRepository
}

func NewStatisticsUsecase(txManager repository.TxManager, userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository, livecommentRepo repository.LivecommentRepository, reactionRepo repository.ReactionRepository, viewerRepo repository.LivestreamViewersHistoryRepository) StatisticsUsecase {
	return &statisticsUsecase{
		txManager:       txManager,
		userRepo:        userRepo,
		livestreamRepo:  livestreamRepo,
		livecommentRepo: livecommentRepo,
		reactionRepo:    reactionRepo,
		viewerRepo:      viewerRepo,
	}
}

func (u *statisticsUsecase) FindUserStatistics(ctx context.Context, username string) (*model.UserStatistics, error) {
	// ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
	// また、現在の合計視聴者数もだす
	var stats *model.UserStatistics
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		user, err := u.userRepo.FindByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// ランク算出
		users, err := u.userRepo.FindAll(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to get users: %w", err)
		}

		var ranking model.UserRanking
		for _, user := range users {
			reactions, err := u.reactionRepo.CountByLivestreamOwnerID(ctx, q, user.ID)
			if err != nil {
				return fmt.Errorf("failed to count reactions: %w", err)
			}

			tips, err := u.livecommentRepo.SumTipByLivestreamOwnerID(ctx, q, user.ID)
			if err != nil {
				return fmt.Errorf("failed to count tips: %w", err)
			}

			score := reactions + tips
			ranking = append(ranking, model.UserRankingEntry{
				Username: user.Name,
				Score:    score,
			})
		}
		sort.Sort(ranking)
		rank := ranking.RankOf(username)

		// リアクション数
		totalReactions, err := u.reactionRepo.CountByLivestreamOwnerName(ctx, q, username)
		if err != nil {
			return fmt.Errorf("failed to count total reactions: %w", err)
		}

		// ライブコメント数、チップ合計
		var totalLivecomments int64
		var totalTip int64
		livestreams, err := u.livestreamRepo.FindAllByUserID(ctx, q, user.ID)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}

		for _, livestream := range livestreams {
			livecomments, err := u.livecommentRepo.FindAllByLivestreamID(ctx, q, livestream.ID)
			if err != nil {
				return fmt.Errorf("failed to get livecomments: %w", err)
			}

			for _, livecomment := range livecomments {
				totalTip += livecomment.Tip
				totalLivecomments++
			}
		}

		// 合計視聴者数
		var viewersCount int64
		for _, livestream := range livestreams {
			cnt, err := u.viewerRepo.CountByLivestreamID(ctx, q, livestream.ID)
			if err != nil {
				return fmt.Errorf("failed to get livestream_view_history: %w", err)
			}
			viewersCount += cnt
		}

		// お気に入り絵文字 (リアクションが無い場合は空文字列)
		favoriteEmoji, err := u.reactionRepo.FindFavoriteEmojiByLivestreamOwnerName(ctx, q, username)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("failed to find favorite emoji: %w", err)
		}

		stats = &model.UserStatistics{
			Rank:              rank,
			ViewersCount:      viewersCount,
			TotalReactions:    totalReactions,
			TotalLivecomments: totalLivecomments,
			TotalTip:          totalTip,
			FavoriteEmoji:     favoriteEmoji,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stats, nil
}
