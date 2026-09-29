package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// reportFiller は f のユーザ、livecommentModels のライブコメント、testLivestreamModel1・2 のライブ配信を引く LivecommentReportFiller を返す。
func (f *livestreamFixture) reportFiller(livecommentModels ...*domain.LivecommentModel) *LivecommentReportFiller {
	return NewLivecommentReportFiller(f.userRepo(), newLivecommentRepositoryWithModels(livecommentModels...), f.userFiller(), f.livecommentFiller(testLivestreamModel1, testLivestreamModel2))
}

// report は、このデータで reportModel を埋めた結果として期待する domain.LivecommentReport を返す。
// livecommentModel, livestreamModel は報告されたライブコメントとそのライブ配信。
func (f *livestreamFixture) report(reportModel *domain.LivecommentReportModel, livecommentModel *domain.LivecommentModel, livestreamModel *domain.LivestreamModel) *domain.LivecommentReport {
	return &domain.LivecommentReport{
		ID:          reportModel.ID,
		Reporter:    f.user(reportModel.UserID),
		Livecomment: *f.livecomment(livecommentModel, livestreamModel),
		CreatedAt:   reportModel.CreatedAt,
	}
}

var (
	testReportModel1 = &domain.LivecommentReportModel{ID: 7, UserID: 42, LivestreamID: 1, LivecommentID: 50, CreatedAt: 400}
	testReportModel2 = &domain.LivecommentReportModel{ID: 8, UserID: 42, LivestreamID: 1, LivecommentID: 51, CreatedAt: 500}
)

func TestLivecommentReportFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

	got, err := f.reportFiller(testLivecommentModel1, testLivecommentModel2).Fill(context.Background(), nil, []*domain.LivecommentReportModel{testReportModel1, testReportModel2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.LivecommentReportID]*domain.LivecommentReport{
		7: f.report(testReportModel1, testLivecommentModel1, testLivestreamModel1),
		8: f.report(testReportModel2, testLivecommentModel2, testLivestreamModel1),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("reports = %+v, want %+v", got, want)
	}
	// 報告したユーザは重複を除いて 1 回だけ引く (コメントしたユーザ・配信者は LivecommentFiller などが別に引く)
	if want := []string{"user 42", "user 43", "user 42", "user 42", "livestream tags 1", "tag 7", "tag 8"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestLivecommentReportFiller_Fill_Errors(t *testing.T) {
	tests := []struct {
		name              string
		modify            func(f *livestreamFixture)
		livecommentModels []*domain.LivecommentModel
		wantMsg           string
	}{
		{
			name:              "reporter not found",
			modify:            func(f *livestreamFixture) { delete(f.users, 42) },
			livecommentModels: []*domain.LivecommentModel{testLivecommentModel1},
			wantMsg:           "failed to get reporter of livecomment report 7: not found",
		},
		{
			name:    "livecomment not found",
			modify:  func(*livestreamFixture) {},
			wantMsg: "failed to get livecomment of livecomment report 7: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			tt.modify(f)
			_, err := f.reportFiller(tt.livecommentModels...).Fill(context.Background(), nil, []*domain.LivecommentReportModel{testReportModel1})
			if !errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
