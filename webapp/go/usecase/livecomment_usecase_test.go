package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestLivecommentUsecase_FindAllByLivestreamID(t *testing.T) {
	limit := int64(5)

	tests := []struct {
		name      string
		limit     *int64
		wantCalls []string
		wantLimit int64
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetailsByLivestreamID"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsByLivestreamIDLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*model.Livecomment{{ID: 1}}
			repo := &fakeLivecommentRepository{livecomments: want}
			u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeLivecommentReportRepository{})

			got, err := u.FindAllByLivestreamID(context.Background(), 10, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
			if !slices.Equal(repo.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", repo.calls, tt.wantCalls)
			}
			if repo.gotLivestreamID != 10 || repo.gotLimit != tt.wantLimit {
				t.Errorf("livestreamID = %d, limit = %d", repo.gotLivestreamID, repo.gotLimit)
			}
		})
	}
}

func TestLivecommentUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, &fakeLivecommentRepository{err: boom}, &fakeLivecommentReportRepository{})

	_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivecommentUsecase_FindAllReportsByLivestreamID(t *testing.T) {
	want := []*model.LivecommentReport{{ID: 1}}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 1}}
	reportRepo := &fakeLivecommentReportRepository{reports: want}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, reportRepo)

	got, err := u.FindAllReportsByLivestreamID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if livestreamRepo.gotID != 10 || reportRepo.gotLivestreamID != 10 {
		t.Errorf("livestream id = %d, report livestream id = %d", livestreamRepo.gotID, reportRepo.gotLivestreamID)
	}
}

func TestLivecommentUsecase_FindAllReportsByLivestreamID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name            string
		livestreamRepo  *fakeLivestreamRepository
		reportRepo      *fakeLivecommentReportRepository
		check           func(t *testing.T, err error)
		wantReportCalls int
	}{
		{
			name:           "not the owner",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}},
			reportRepo:     &fakeLivecommentReportRepository{},
			check: func(t *testing.T, err error) {
				if !errors.Is(err, ErrNotLivestreamOwner) {
					t.Errorf("err = %v, want ErrNotLivestreamOwner", err)
				}
			},
		},
		{
			// ライブ配信が無い場合は 404 用のエラーには変換しない (移行前と同じく 500)
			name:           "livestream not found",
			livestreamRepo: &fakeLivestreamRepository{err: repository.ErrNotFound},
			reportRepo:     &fakeLivecommentReportRepository{},
			check: func(t *testing.T, err error) {
				if err == nil || errors.Is(err, ErrLivestreamNotFound) || errors.Is(err, ErrNotLivestreamOwner) {
					t.Errorf("err = %v, want a plain error", err)
				}
			},
		},
		{
			name:            "report repository error",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 1}},
			reportRepo:      &fakeLivecommentReportRepository{err: boom},
			wantReportCalls: 1,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, boom) {
					t.Errorf("err = %v, want %v", err, boom)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, &fakeLivecommentRepository{}, tt.reportRepo)
			_, err := u.FindAllReportsByLivestreamID(context.Background(), 1, 10)
			tt.check(t, err)
			if len(tt.reportRepo.calls) != tt.wantReportCalls {
				t.Errorf("report calls = %v", tt.reportRepo.calls)
			}
		})
	}
}
