package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type LivecommentReportRepository interface {
	// FindByID は報告が存在しない場合 ErrNotFound を返す。
	FindByID(ctx context.Context, q Querier, id domain.LivecommentReportID) (*domain.LivecommentReport, error)
	// FindAllByLivestreamID は指定したライブ配信へのライブコメントの報告を返す (順序は不定)。
	FindAllByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReport, error)
	// CountByLivestreamID は指定したライブ配信へのライブコメントの報告数を返す。
	CountByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) (int64, error)
	// Create はライブコメントの報告を登録し、その ID を返す。
	Create(ctx context.Context, q Querier, report *domain.LivecommentReport) (domain.LivecommentReportID, error)
}
