package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type LivecommentReportRepository interface {
	// FindAllWithDetailsByLivestreamID は指定したライブ配信へのライブコメントの報告を返す。
	FindAllWithDetailsByLivestreamID(ctx context.Context, q Querier, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error)
}
