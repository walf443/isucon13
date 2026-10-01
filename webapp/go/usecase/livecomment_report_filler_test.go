package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestLivecommentReportFiller_Fill(t *testing.T) {
	f := testDetailFixture()

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
	if want := []string{"users [42]", "users [43 42]", "users [42]", "livestream tags [1]", "tags [7 8]"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestLivecommentReportFiller_Fill_Errors(t *testing.T) {
	tests := []struct {
		name         string
		modify       func(f *detailFixture)
		livecomments []*domain.Livecomment
		wantMsg      string
	}{
		{
			name:         "reporter not found",
			modify:       func(f *detailFixture) { delete(f.users, 42) },
			livecomments: []*domain.Livecomment{testLivecomment1},
			wantMsg:      "failed to get reporter of livecomment report 7: not found",
		},
		{
			name:    "livecomment not found",
			modify:  func(*detailFixture) {},
			wantMsg: "failed to get livecomment of livecomment report 7: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testDetailFixture()
			tt.modify(f)
			_, err := f.reportFiller(tt.livecomments...).Fill(context.Background(), nil, []*domain.LivecommentReport{testReport1})
			if !errors.Is(err, errMissingDetail) || errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

func TestLivecommentReportFiller_Fill_RepositoryErrors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		modify  func(filler *LivecommentReportFiller)
		wantMsg string
	}{
		{
			name: "get reporters fails",
			modify: func(filler *LivecommentReportFiller) {
				filler.userRepo = &fakeUserRepository{
					findAllByIDs: func(context.Context, repository.Querier, []domain.UserID) ([]*domain.User, error) { return nil, boom },
				}
			},
			wantMsg: "failed to get reporters: boom",
		},
		{
			name: "get livecomments fails",
			modify: func(filler *LivecommentReportFiller) {
				filler.livecommentRepo = &fakeLivecommentRepository{
					findAllByIDs: func(context.Context, repository.Querier, []domain.LivecommentID) ([]*domain.Livecomment, error) {
						return nil, boom
					},
				}
			},
			wantMsg: "failed to get livecomments: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filler := testDetailFixture().reportFiller(testLivecomment1)
			tt.modify(filler)
			_, err := filler.Fill(context.Background(), nil, []*domain.LivecommentReport{testReport1})
			if !errors.Is(err, boom) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
