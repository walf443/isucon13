package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// reportFiller は f のユーザ、livecomments のライブコメント、testLivestream1・2 のライブ配信を引く LivecommentReportFiller を返す。
func (f *livestreamFixture) reportFiller(livecomments ...*domain.Livecomment) *LivecommentReportFiller {
	return NewLivecommentReportFiller(f.userRepo(), newLivecommentRepositoryWithLivecomments(livecomments...), f.userFiller(), f.livecommentFiller(testLivestream1, testLivestream2))
}

// report は、このデータで report を埋めた結果として期待する domain.LivecommentReportDetail を返す。
// livecomment, livestream は報告されたライブコメントとそのライブ配信。
func (f *livestreamFixture) report(report *domain.LivecommentReport, livecomment *domain.Livecomment, livestream *domain.Livestream) *domain.LivecommentReportDetail {
	return &domain.LivecommentReportDetail{
		ID:          report.ID,
		Reporter:    f.user(report.UserID),
		Livecomment: *f.livecomment(livecomment, livestream),
		CreatedAt:   report.CreatedAt,
	}
}

var (
	testReport1 = &domain.LivecommentReport{ID: 7, UserID: 42, LivestreamID: 1, LivecommentID: 50, CreatedAt: 400}
	testReport2 = &domain.LivecommentReport{ID: 8, UserID: 42, LivestreamID: 1, LivecommentID: 51, CreatedAt: 500}
)

func TestLivecommentReportFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

	got, err := f.reportFiller(testLivecomment1, testLivecomment2).Fill(context.Background(), nil, []*domain.LivecommentReport{testReport1, testReport2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.LivecommentReportID]*domain.LivecommentReportDetail{
		7: f.report(testReport1, testLivecomment1, testLivestream1),
		8: f.report(testReport2, testLivecomment2, testLivestream1),
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
		name         string
		modify       func(f *livestreamFixture)
		livecomments []*domain.Livecomment
		wantMsg      string
	}{
		{
			name:         "reporter not found",
			modify:       func(f *livestreamFixture) { delete(f.users, 42) },
			livecomments: []*domain.Livecomment{testLivecomment1},
			wantMsg:      "failed to get reporter of livecomment report 7: not found",
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
			_, err := f.reportFiller(tt.livecomments...).Fill(context.Background(), nil, []*domain.LivecommentReport{testReport1})
			if !errors.Is(err, errMissingDetail) || errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
