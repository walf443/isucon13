package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type NGWordUsecase interface {
	// FindAllByLivestreamID は userID のユーザがライブ配信に登録した NG ワードを、作成日時の降順で返す。
	FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error)
	// Moderate は配信者がライブ配信に NG ワードを登録し、NG ワードに当たる過去のライブコメントを削除する。
	// 登録した NG ワードの ID を返す。
	// userID のユーザが配信者でない場合 (ライブ配信が存在しない場合を含む) ErrNotLivestreamOwner を返す。
	Moderate(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, word string) (model.NGWordID, error)
}

type ngWordUsecase struct {
	txManager       repository.TxManager
	livestreamRepo  repository.LivestreamRepository
	livecommentRepo repository.LivecommentRepository
	ngWordRepo      repository.NGWordRepository
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func NewNGWordUsecase(txManager repository.TxManager, livestreamRepo repository.LivestreamRepository, livecommentRepo repository.LivecommentRepository, ngWordRepo repository.NGWordRepository) NGWordUsecase {
	return &ngWordUsecase{
		txManager:       txManager,
		livestreamRepo:  livestreamRepo,
		livecommentRepo: livecommentRepo,
		ngWordRepo:      ngWordRepo,
		now:             time.Now,
	}
}

func (u *ngWordUsecase) FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	var ngWords []*model.NGWordModel
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		ngWords, err = u.ngWordRepo.FindAllByUserIDAndLivestreamID(ctx, q, userID, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to get NG words: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ngWords, nil
}

func (u *ngWordUsecase) Moderate(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, word string) (model.NGWordID, error) {
	var wordID model.NGWordID
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		// 配信者自身の配信に対するmoderateなのかを検証
		ownedLivestreams, err := u.livestreamRepo.FindAllByIDAndUserID(ctx, q, livestreamID, userID)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		if len(ownedLivestreams) == 0 {
			return ErrNotLivestreamOwner
		}

		wordID, err = u.ngWordRepo.Create(ctx, q, &model.NGWordModel{
			UserID:       userID,
			LivestreamID: livestreamID,
			Word:         word,
			CreatedAt:    u.now().Unix(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert new NG word: %w", err)
		}

		ngWords, err := u.ngWordRepo.FindAllByLivestreamID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to get NG words: %w", err)
		}

		// NGワードにヒットする過去の投稿も全削除する
		for _, ngWord := range ngWords {
			if err := u.livecommentRepo.DeleteAllByLivestreamIDMatchingNGWord(ctx, q, livestreamID, ngWord.Word); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return wordID, nil
}
