package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestLivecommentReportUsecase_FindAllByLivestreamID(t *testing.T) {
	want := []*model.LivecommentReport{{ID: 1}}
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 10, &model.LivestreamModel{ID: 10, UserID: 1}, nil)
	reportRepo := &fakeLivecommentReportRepository{
		findAllWithDetailsByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
			if livestreamID != 10 {
				t.Errorf("report livestream id = %d, want 10", livestreamID)
			}
			return want, nil
		},
	}
	u := NewLivecommentReportUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, reportRepo)

	got, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivecommentReportUsecase_FindAllByLivestreamID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// livestream, livestreamErr はライブ配信 10 を引いた結果
		livestream    *model.LivestreamModel
		livestreamErr error
		// reportsErr は報告の取得が返すエラー
		reportsErr      error
		check           func(t *testing.T, err error)
		wantReportCalls int
	}{
		{
			name:       "not the owner",
			livestream: &model.LivestreamModel{ID: 10, UserID: 2},
			check: func(t *testing.T, err error) {
				if !errors.Is(err, ErrNotLivestreamOwner) {
					t.Errorf("err = %v, want ErrNotLivestreamOwner", err)
				}
			},
		},
		{
			// ライブ配信が無い場合は 404 用のエラーには変換しない (移行前と同じく 500)
			name:          "livestream not found",
			livestreamErr: repository.ErrNotFound,
			check: func(t *testing.T, err error) {
				if err == nil || errors.Is(err, ErrLivestreamNotFound) || errors.Is(err, ErrNotLivestreamOwner) {
					t.Errorf("err = %v, want a plain error", err)
				}
			},
		},
		{
			name:            "report repository error",
			livestream:      &model.LivestreamModel{ID: 10, UserID: 1},
			reportsErr:      boom,
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
			reportCalls := 0
			reportRepo := &fakeLivecommentReportRepository{
				findAllWithDetailsByLivestreamID: func(context.Context, repository.Querier, model.LivestreamID) ([]*model.LivecommentReport, error) {
					reportCalls++
					return nil, tt.reportsErr
				},
			}
			u := NewLivecommentReportUsecase(&fakeTxManager{}, newLivestreamRepositoryFindingByID(t, 10, tt.livestream, tt.livestreamErr), &fakeLivecommentRepository{}, reportRepo)
			_, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
			tt.check(t, err)
			if reportCalls != tt.wantReportCalls {
				t.Errorf("report calls = %d, want %d", reportCalls, tt.wantReportCalls)
			}
		})
	}
}

// newReportRepositoryForReport は Report で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeLivecommentReportRepository を返す。
// 登録した報告は created に取り出す。
func newReportRepositoryForReport(t *testing.T, calls *[]string, created **model.LivecommentReportModel, report *model.LivecommentReport, createErr, fillErr error) *fakeLivecommentReportRepository {
	return &fakeLivecommentReportRepository{
		create: func(_ context.Context, _ repository.Querier, r *model.LivecommentReportModel) (model.LivecommentReportID, error) {
			*calls = append(*calls, "Create")
			*created = r
			return 7, createErr
		},
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.LivecommentReportID) (*model.LivecommentReport, error) {
			*calls = append(*calls, "FindWithDetailsByID")
			if id != 7 {
				t.Errorf("re-read id = %d, want 7", id)
			}
			return report, fillErr
		},
	}
}

func TestLivecommentReportUsecase_Create(t *testing.T) {
	want := &model.LivecommentReport{ID: 7}
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 10, &model.LivestreamModel{ID: 10, UserID: 2}, nil)
	livecommentFinds := 0
	livecommentRepo := &fakeLivecommentRepository{
		findByID: func(_ context.Context, _ repository.Querier, id model.LivecommentID) (*model.LivecommentModel, error) {
			livecommentFinds++
			if id != 50 {
				t.Errorf("livecomment id = %d, want 50", id)
			}
			return &model.LivecommentModel{ID: 50, LivestreamID: 10}, nil
		},
	}
	var reportCalls []string
	var created *model.LivecommentReportModel
	reportRepo := newReportRepositoryForReport(t, &reportCalls, &created, want, nil, nil)
	u := NewLivecommentReportUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, reportRepo).(*livecommentReportUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 3, 10, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if livecommentFinds != 1 {
		t.Errorf("livecomment finds = %d, want 1", livecommentFinds)
	}
	wantCreated := model.LivecommentReportModel{UserID: 3, LivestreamID: 10, LivecommentID: 50, CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
	if want := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(reportCalls, want) {
		t.Errorf("report calls = %v, want %v", reportCalls, want)
	}
}

func TestLivecommentReportUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")
	livestream := &model.LivestreamModel{ID: 10, UserID: 2}
	livecomment := &model.LivecommentModel{ID: 50, LivestreamID: 10}

	tests := []struct {
		name string
		// livestream, livestreamErr はライブ配信 10 を引いた結果
		livestream    *model.LivestreamModel
		livestreamErr error
		// livecomment, livecommentErr はライブコメントの存在確認の結果
		livecomment    *model.LivecommentModel
		livecommentErr error
		// reportCreateErr, reportFillErr は報告の登録・取り直しが返すエラー
		reportCreateErr error
		reportFillErr   error
		wantErr         error
		wantReportCalls []string
	}{
		{
			name:          "livestream not found",
			livestreamErr: repository.ErrNotFound,
			wantErr:       ErrLivestreamNotFound,
		},
		{
			name:          "livestream repository error",
			livestreamErr: boom,
			wantErr:       boom,
		},
		{
			name:           "livecomment not found",
			livestream:     livestream,
			livecommentErr: repository.ErrNotFound,
			wantErr:        ErrLivecommentNotFound,
		},
		{
			name:           "livecomment repository error",
			livestream:     livestream,
			livecommentErr: boom,
			wantErr:        boom,
		},
		{
			name:            "create fails",
			livestream:      livestream,
			livecomment:     livecomment,
			reportCreateErr: boom,
			wantErr:         boom,
			wantReportCalls: []string{"Create"},
		},
		{
			name:            "fill fails",
			livestream:      livestream,
			livecomment:     livecomment,
			reportFillErr:   boom,
			wantErr:         boom,
			wantReportCalls: []string{"Create", "FindWithDetailsByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reportCalls []string
			var created *model.LivecommentReportModel
			reportRepo := newReportRepositoryForReport(t, &reportCalls, &created, nil, tt.reportCreateErr, tt.reportFillErr)
			livecommentRepo := &fakeLivecommentRepository{
				findByID: func(_ context.Context, _ repository.Querier, id model.LivecommentID) (*model.LivecommentModel, error) {
					if id != 50 {
						t.Errorf("livecomment id = %d, want 50", id)
					}
					return tt.livecomment, tt.livecommentErr
				},
			}
			u := NewLivecommentReportUsecase(&fakeTxManager{}, newLivestreamRepositoryFindingByID(t, 10, tt.livestream, tt.livestreamErr), livecommentRepo, reportRepo)
			_, err := u.Create(context.Background(), 3, 10, 50)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			// ライブ配信やライブコメントが無い場合は報告を登録しない
			if !slices.Equal(reportCalls, tt.wantReportCalls) {
				t.Errorf("report calls = %v, want %v", reportCalls, tt.wantReportCalls)
			}
		})
	}
}
