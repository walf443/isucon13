package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type LivecommentRepository interface {
	// FindWithDetailsByID はユーザ・ライブ配信を含めたライブコメントを返す。
	// ライブコメントが存在しない場合 ErrNotFound を返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id domain.LivecommentID) (*domain.Livecomment, error)
	// FindAllWithDetailsByLivestreamID は指定したライブ配信へのライブコメントを、作成日時の降順で返す。
	FindAllWithDetailsByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) ([]*domain.Livecomment, error)
	// FindAllWithDetailsByLivestreamIDLimited は指定したライブ配信へのライブコメントを、作成日時の降順で最大 limit 件返す。
	FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Livecomment, error)
	// FindAllByLivestreamID は指定したライブ配信へのライブコメントを返す (順序は不定)。
	FindAllByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentModel, error)
	// SumTip は全てのライブコメントのチップ合計を返す。
	SumTip(ctx context.Context, q Querier) (int64, error)
	// SumTipByLivestreamID は指定したライブ配信へのライブコメントのチップ合計を返す。
	SumTipByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) (int64, error)
	// MaxTipByLivestreamID は指定したライブ配信へのライブコメントのチップの最大値を返す。ライブコメントが無い場合は 0 を返す。
	MaxTipByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) (int64, error)
	// SumTipByLivestreamOwnerID は指定したユーザが配信者のライブ配信へのライブコメントのチップ合計を返す。
	SumTipByLivestreamOwnerID(ctx context.Context, q Querier, userID domain.UserID) (int64, error)
	// FindByID はライブコメントが存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id domain.LivecommentID) (*domain.LivecommentModel, error)
	// Create はライブコメントを登録し、その ID を返す。
	Create(ctx context.Context, q Querier, livecomment *domain.LivecommentModel) (domain.LivecommentID, error)
	// DeleteAllByLivestreamIDMatchingNGWord は指定したライブ配信へのライブコメントのうち、NG ワード word に当たるものを削除する。
	// 判定は NGWordRepository.Matches と同じく MySQL の LIKE '%word%' で行う。
	DeleteAllByLivestreamIDMatchingNGWord(ctx context.Context, q Querier, livestreamID domain.LivestreamID, word string) error
}
