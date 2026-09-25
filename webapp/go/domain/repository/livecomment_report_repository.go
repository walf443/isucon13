package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type LivecommentReportRepository interface {
	// FindAllWithDetailsByLivestreamID は指定したライブ配信へのライブコメントの報告を返す。
	FindAllWithDetailsByLivestreamID(ctx context.Context, q Querier, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error)
	// FindWithDetailsByID は報告したユーザ・報告されたライブコメントを含めた報告を返す。
	// 報告が存在しない場合 ErrNotFound を返す。報告したユーザやライブコメントが欠けている場合は ErrNotFound ではないエラーを返す。
	FindWithDetailsByID(ctx context.Context, q Querier, id model.LivecommentReportID) (*model.LivecommentReport, error)
	// Create はライブコメントの報告を登録し、その ID を返す。
	Create(ctx context.Context, q Querier, report *model.LivecommentReportModel) (model.LivecommentReportID, error)
}
