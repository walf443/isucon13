package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type ReactionUsecase interface {
	// FindAllByLivestreamID は指定したライブ配信へのリアクションを作成日時の降順で返す。
	// limit が nil でなければ最大 *limit 件に絞る。
	FindAllByLivestreamID(ctx context.Context, livestreamID domain.LivestreamID, limit *domain.Limit) ([]*domain.ReactionDetail, error)
	// Create はリアクションを登録し、ユーザ・ライブ配信を含めて返す。
	Create(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID, emojiName string) (*domain.ReactionDetail, error)
}

type reactionUsecase struct {
	txManager      repository.TxManager
	reactionRepo   repository.ReactionRepository
	reactionFiller *ReactionFiller
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func NewReactionUsecase(txManager repository.TxManager, reactionRepo repository.ReactionRepository, reactionFiller *ReactionFiller) ReactionUsecase {
	return &reactionUsecase{txManager: txManager, reactionRepo: reactionRepo, reactionFiller: reactionFiller, now: time.Now}
}

func (u *reactionUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID domain.LivestreamID, limit *domain.Limit) ([]*domain.ReactionDetail, error) {
	var reactionDetails []*domain.ReactionDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var reactions []*domain.Reaction
		var err error
		if limit == nil {
			reactions, err = u.reactionRepo.FindAllByLivestreamIDOrderByCreatedAtDesc(ctx, q, livestreamID)
		} else {
			reactions, err = u.reactionRepo.FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx, q, livestreamID, *limit)
		}
		if err != nil {
			return fmt.Errorf("failed to get reactions: %w", err)
		}

		filled, err := u.reactionFiller.Fill(ctx, q, reactions)
		if err != nil {
			return fmt.Errorf("failed to get reactions: %w", err)
		}
		reactionDetails = orderedBy(reactions, filled)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reactionDetails, nil
}

func (u *reactionUsecase) Create(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID, emojiName string) (*domain.ReactionDetail, error) {
	var reactionDetail *domain.ReactionDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		reactionID, err := u.reactionRepo.Create(ctx, q, &domain.Reaction{
			UserID:       userID,
			LivestreamID: livestreamID,
			EmojiName:    emojiName,
			CreatedAt:    u.now().Unix(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert reaction: %w", err)
		}

		reaction, err := u.reactionRepo.FindByID(ctx, q, reactionID)
		if err != nil {
			return fmt.Errorf("failed to fill reaction: %w", err)
		}
		reactionDetails, err := u.reactionFiller.Fill(ctx, q, []*domain.Reaction{reaction})
		if err != nil {
			return fmt.Errorf("failed to fill reaction: %w", err)
		}
		reactionDetail = reactionDetails[reactionID]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reactionDetail, nil
}
