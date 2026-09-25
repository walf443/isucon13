package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeLivecommentReportRepository struct {
	reportCount    int64
	reportCountErr error

	report    *model.LivecommentReport
	reports   []*model.LivecommentReport
	err       error
	createID  model.LivecommentReportID
	createErr error

	calls           []string
	gotID           model.LivecommentReportID
	gotLivestreamID model.LivestreamID
	gotCreated      *model.LivecommentReportModel
}

func (r *fakeLivecommentReportRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	r.gotLivestreamID = livestreamID
	return r.reportCount, r.reportCountErr
}

func (r *fakeLivecommentReportRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivecommentReportID) (*model.LivecommentReport, error) {
	r.calls = append(r.calls, "FindWithDetailsByID")
	r.gotID = id
	return r.report, r.err
}

func (r *fakeLivecommentReportRepository) Create(ctx context.Context, q repository.Querier, report *model.LivecommentReportModel) (model.LivecommentReportID, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = report
	return r.createID, r.createErr
}

func (r *fakeLivecommentReportRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamID")
	r.gotLivestreamID = livestreamID
	return r.reports, r.err
}
