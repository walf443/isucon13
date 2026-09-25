package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type ReactionUsecase interface {
	// FindAllByLivestreamID は指定したライブ配信へのリアクションを作成日時の降順で返す。
	// limit が nil でなければ最大 *limit 件に絞る。
	FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *model.Limit) ([]*model.Reaction, error)
	// Create はリアクションを登録し、ユーザ・ライブ配信を含めて返す。
	Create(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, emojiName string) (*model.Reaction, error)
}

type reactionUsecase struct {
	txManager    repository.TxManager
	reactionRepo repository.ReactionRepository
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func NewReactionUsecase(txManager repository.TxManager, reactionRepo repository.ReactionRepository) ReactionUsecase {
	return &reactionUsecase{txManager: txManager, reactionRepo: reactionRepo, now: time.Now}
}

func (u *reactionUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *model.Limit) ([]*model.Reaction, error) {
	var reactions []*model.Reaction
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		if limit == nil {
			reactions, err = u.reactionRepo.FindAllWithDetailsByLivestreamID(ctx, q, livestreamID)
		} else {
			reactions, err = u.reactionRepo.FindAllWithDetailsByLivestreamIDLimited(ctx, q, livestreamID, *limit)
		}
		if err != nil {
			return fmt.Errorf("failed to get reactions: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

func (u *reactionUsecase) Create(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, emojiName string) (*model.Reaction, error) {
	var reaction *model.Reaction
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		reactionID, err := u.reactionRepo.Create(ctx, q, &model.ReactionModel{
			UserID:       userID,
			LivestreamID: livestreamID,
			EmojiName:    emojiName,
			CreatedAt:    u.now().Unix(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert reaction: %w", err)
		}

		reaction, err = u.reactionRepo.FindWithDetailsByID(ctx, q, reactionID)
		if err != nil {
			return fmt.Errorf("failed to fill reaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reaction, nil
}
