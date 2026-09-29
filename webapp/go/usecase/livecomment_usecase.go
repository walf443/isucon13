package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type LivecommentUsecase interface {
	// FindAllByLivestreamID は指定したライブ配信へのライブコメントを作成日時の降順で返す。
	// limit が nil でなければ最大 *limit 件に絞る。
	FindAllByLivestreamID(ctx context.Context, livestreamID domain.LivestreamID, limit *domain.Limit) ([]*domain.LivecommentDetail, error)
	// Create はライブコメントを投稿し、ユーザ・ライブ配信を含めて返す。
	// ライブ配信が存在しない場合 ErrLivestreamNotFound、配信者の NG ワードに当たった場合 ErrSpamLivecomment を返す。
	Create(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID, comment string, tip int64) (*domain.LivecommentDetail, error)
}

type livecommentUsecase struct {
	txManager         repository.TxManager
	livestreamRepo    repository.LivestreamRepository
	livecommentRepo   repository.LivecommentRepository
	ngWordRepo        repository.NGWordRepository
	livecommentFiller *LivecommentFiller
	logger            Logger
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func NewLivecommentUsecase(txManager repository.TxManager, livestreamRepo repository.LivestreamRepository, livecommentRepo repository.LivecommentRepository, ngWordRepo repository.NGWordRepository, livecommentFiller *LivecommentFiller, logger Logger) LivecommentUsecase {
	return &livecommentUsecase{
		txManager:         txManager,
		livestreamRepo:    livestreamRepo,
		livecommentRepo:   livecommentRepo,
		ngWordRepo:        ngWordRepo,
		livecommentFiller: livecommentFiller,
		logger:            logger,
		now:               time.Now,
	}
}

func (u *livecommentUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID domain.LivestreamID, limit *domain.Limit) ([]*domain.LivecommentDetail, error) {
	var livecommentDetails []*domain.LivecommentDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var livecomments []*domain.Livecomment
		var err error
		if limit == nil {
			livecomments, err = u.livecommentRepo.FindAllByLivestreamIDOrderByCreatedAtDesc(ctx, q, livestreamID)
		} else {
			livecomments, err = u.livecommentRepo.FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx, q, livestreamID, *limit)
		}
		if err != nil {
			return fmt.Errorf("failed to get livecomments: %w", err)
		}

		filled, err := u.livecommentFiller.Fill(ctx, q, livecomments)
		if err != nil {
			return fmt.Errorf("failed to get livecomments: %w", err)
		}
		livecommentDetails = orderedBy(livecomments, filled)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livecommentDetails, nil
}

func (u *livecommentUsecase) Create(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID, comment string, tip int64) (*domain.LivecommentDetail, error) {
	var livecommentDetail *domain.LivecommentDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		livestream, err := u.livestreamRepo.FindByID(ctx, q, livestreamID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}

		// スパム判定 (配信者が登録した NG ワードに当たるか)
		ngWords, err := u.ngWordRepo.FindAllByUserIDAndLivestreamID(ctx, q, livestream.UserID, livestream.ID)
		if err != nil {
			return fmt.Errorf("failed to get NG words: %w", err)
		}
		for _, ngWord := range ngWords {
			hit, err := u.ngWordRepo.Matches(ctx, q, comment, ngWord.Word)
			if err != nil {
				return fmt.Errorf("failed to get hitspam: %w", err)
			}
			// 移行前は SQL の COUNT(*) (常に 0 か 1) をそのままログに出していたので、同じ形式で出す
			hitSpam := 0
			if hit {
				hitSpam = 1
			}
			u.logger.Infof("[hitSpam=%d] comment = %s", hitSpam, comment)
			if hit {
				return ErrSpamLivecomment
			}
		}

		livecommentID, err := u.livecommentRepo.Create(ctx, q, &domain.Livecomment{
			UserID:       userID,
			LivestreamID: livestreamID,
			Comment:      comment,
			Tip:          tip,
			CreatedAt:    u.now().Unix(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert livecomment: %w", err)
		}

		livecomment, err := u.livecommentRepo.FindByID(ctx, q, livecommentID)
		if err != nil {
			return fmt.Errorf("failed to fill livecomment: %w", err)
		}
		livecommentDetails, err := u.livecommentFiller.Fill(ctx, q, []*domain.Livecomment{livecomment})
		if err != nil {
			return fmt.Errorf("failed to fill livecomment: %w", err)
		}
		livecommentDetail = livecommentDetails[livecommentID]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livecommentDetail, nil
}
