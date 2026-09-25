package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type StatisticsUsecase interface {
	// FindUserStatistics は指定したユーザの配信者としての統計情報を返す。
	// ユーザが存在しない場合 ErrUserNotFound を返す。
	FindUserStatistics(ctx context.Context, username string) (*domain.UserStatistics, error)
	// FindLivestreamStatistics は指定したライブ配信の統計情報を返す。
	// ライブ配信が存在しない場合 ErrLivestreamNotFound を返す。
	FindLivestreamStatistics(ctx context.Context, livestreamID domain.LivestreamID) (*domain.LivestreamStatistics, error)
}

type statisticsUsecase struct {
	txManager       repository.TxManager
	userRepo        repository.UserRepository
	livestreamRepo  repository.LivestreamRepository
	livecommentRepo repository.LivecommentRepository
	reactionRepo    repository.ReactionRepository
	viewerRepo      repository.LivestreamViewersHistoryRepository
	reportRepo      repository.LivecommentReportRepository
}

func NewStatisticsUsecase(txManager repository.TxManager, userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository, livecommentRepo repository.LivecommentRepository, reactionRepo repository.ReactionRepository, viewerRepo repository.LivestreamViewersHistoryRepository, reportRepo repository.LivecommentReportRepository) StatisticsUsecase {
	return &statisticsUsecase{
		txManager:       txManager,
		userRepo:        userRepo,
		livestreamRepo:  livestreamRepo,
		livecommentRepo: livecommentRepo,
		reactionRepo:    reactionRepo,
		viewerRepo:      viewerRepo,
		reportRepo:      reportRepo,
	}
}

func (u *statisticsUsecase) FindUserStatistics(ctx context.Context, username string) (*domain.UserStatistics, error) {
	// ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
	// また、現在の合計視聴者数もだす
	var stats *domain.UserStatistics
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

		var ranking domain.UserRanking
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
			ranking = append(ranking, domain.UserRankingEntry{
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

		stats = &domain.UserStatistics{
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

func (u *statisticsUsecase) FindLivestreamStatistics(ctx context.Context, livestreamID domain.LivestreamID) (*domain.LivestreamStatistics, error) {
	var stats *domain.LivestreamStatistics
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		_, err := u.livestreamRepo.FindByID(ctx, q, livestreamID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}

		livestreams, err := u.livestreamRepo.FindAll(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}

		// ランク算出
		var ranking domain.LivestreamRanking
		for _, livestream := range livestreams {
			reactions, err := u.reactionRepo.CountByLivestreamID(ctx, q, livestream.ID)
			if err != nil {
				return fmt.Errorf("failed to count reactions: %w", err)
			}

			totalTips, err := u.livecommentRepo.SumTipByLivestreamID(ctx, q, livestream.ID)
			if err != nil {
				return fmt.Errorf("failed to count tips: %w", err)
			}

			score := reactions + totalTips
			ranking = append(ranking, domain.LivestreamRankingEntry{
				LivestreamID: livestream.ID,
				Score:        score,
			})
		}
		sort.Sort(ranking)
		rank := ranking.RankOf(livestreamID)

		// 視聴者数算出
		viewersCount, err := u.viewerRepo.CountViewersByLivestreamID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to count livestream viewers: %w", err)
		}

		// 最大チップ額
		maxTip, err := u.livecommentRepo.MaxTipByLivestreamID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to find maximum tip livecomment: %w", err)
		}

		// リアクション数
		totalReactions, err := u.reactionRepo.CountTotalByLivestreamID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to count total reactions: %w", err)
		}

		// スパム報告数
		totalReports, err := u.reportRepo.CountByLivestreamID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to count total spam reports: %w", err)
		}

		stats = &domain.LivestreamStatistics{
			Rank:           rank,
			ViewersCount:   viewersCount,
			TotalReactions: totalReactions,
			TotalReports:   totalReports,
			MaxTip:         maxTip,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stats, nil
}
