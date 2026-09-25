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

func TestLivecommentUsecase_FindAllByLivestreamID(t *testing.T) {
	limit := model.Limit(5)

	tests := []struct {
		name      string
		limit     *model.Limit
		wantCalls []string
		wantLimit model.Limit
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetailsByLivestreamID"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsByLivestreamIDLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*model.Livecomment{{ID: 1}}
			repo := &fakeLivecommentRepository{livecomments: want}
			u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeLivecommentReportRepository{}, &fakeNGWordRepository{}, &fakeLogger{})

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
	u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, &fakeLivecommentRepository{err: boom}, &fakeLivecommentReportRepository{}, &fakeNGWordRepository{}, &fakeLogger{})

	_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivecommentUsecase_FindAllReportsByLivestreamID(t *testing.T) {
	want := []*model.LivecommentReport{{ID: 1}}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 1}}
	reportRepo := &fakeLivecommentReportRepository{reports: want}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, reportRepo, &fakeNGWordRepository{}, &fakeLogger{})

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
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, &fakeLivecommentRepository{}, tt.reportRepo, &fakeNGWordRepository{}, &fakeLogger{})
			_, err := u.FindAllReportsByLivestreamID(context.Background(), 1, 10)
			tt.check(t, err)
			if len(tt.reportRepo.calls) != tt.wantReportCalls {
				t.Errorf("report calls = %v", tt.reportRepo.calls)
			}
		})
	}
}

func TestLivecommentUsecase_Create(t *testing.T) {
	want := &model.Livecomment{ID: 100, Comment: "hello"}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}}
	livecommentRepo := &fakeLivecommentRepository{createID: 100, livecomment: want}
	ngWordRepo := &fakeNGWordRepository{ngWords: []*model.NGWordModel{{Word: "bad"}, {Word: "evil"}}}
	logger := &fakeLogger{}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, &fakeLivecommentReportRepository{}, ngWordRepo, logger).(*livecommentUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 1, 10, "hello", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// NG ワードは投稿者ではなく配信者のものを使う
	if ngWordRepo.gotUserID != 2 || ngWordRepo.gotLivestreamID != 10 {
		t.Errorf("NG words of userID = %d, livestreamID = %d, want 2, 10", ngWordRepo.gotUserID, ngWordRepo.gotLivestreamID)
	}
	if want := []string{"bad", "evil"}; !slices.Equal(ngWordRepo.matchedWords, want) {
		t.Errorf("matched words = %v, want %v", ngWordRepo.matchedWords, want)
	}
	// NG ワードごとに判定結果をログに出す (移行前と同じ形式)
	if want := []string{"[hitSpam=0] comment = hello", "[hitSpam=0] comment = hello"}; !slices.Equal(logger.lines, want) {
		t.Errorf("log lines = %q, want %q", logger.lines, want)
	}
	if want := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(livecommentRepo.calls, want) {
		t.Errorf("calls = %v, want %v", livecommentRepo.calls, want)
	}
	wantCreated := model.LivecommentModel{UserID: 1, LivestreamID: 10, Comment: "hello", Tip: 500, CreatedAt: 1700000000}
	if *livecommentRepo.gotCreated != wantCreated {
		t.Errorf("created = %+v, want %+v", *livecommentRepo.gotCreated, wantCreated)
	}
	if livecommentRepo.gotID != 100 {
		t.Errorf("re-read id = %d, want 100", livecommentRepo.gotID)
	}
}

func TestLivecommentUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")
	owned := &model.LivestreamModel{ID: 10, UserID: 2}

	tests := []struct {
		name            string
		livestreamRepo  *fakeLivestreamRepository
		livecommentRepo *fakeLivecommentRepository
		ngWordRepo      *fakeNGWordRepository
		wantErr         error
		wantCalls       []string
	}{
		{
			name:            "livestream not found",
			livestreamRepo:  &fakeLivestreamRepository{err: repository.ErrNotFound},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{},
			wantErr:         ErrLivestreamNotFound,
		},
		{
			name:            "spam",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: owned},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{ngWords: []*model.NGWordModel{{Word: "ok"}, {Word: "bad"}}, hitWords: []string{"bad"}},
			wantErr:         ErrSpamLivecomment,
		},
		{
			name:            "NG words repository error",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: owned},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{err: boom},
			wantErr:         boom,
		},
		{
			name:            "match error",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: owned},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{ngWords: []*model.NGWordModel{{Word: "bad"}}, matchErr: boom},
			wantErr:         boom,
		},
		{
			name:            "create fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: owned},
			livecommentRepo: &fakeLivecommentRepository{createErr: boom},
			ngWordRepo:      &fakeNGWordRepository{},
			wantErr:         boom,
			wantCalls:       []string{"Create"},
		},
		{
			name:            "fill fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: owned},
			livecommentRepo: &fakeLivecommentRepository{createID: 100, err: boom},
			ngWordRepo:      &fakeNGWordRepository{},
			wantErr:         boom,
			wantCalls:       []string{"Create", "FindWithDetailsByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, tt.livecommentRepo, &fakeLivecommentReportRepository{}, tt.ngWordRepo, &fakeLogger{})
			_, err := u.Create(context.Background(), 1, 10, "this is bad", 0)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			// スパム判定などで失敗した場合はライブコメントを登録しない
			if !slices.Equal(tt.livecommentRepo.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", tt.livecommentRepo.calls, tt.wantCalls)
			}
		})
	}
}

func TestLivecommentUsecase_Create_LogsHitSpam(t *testing.T) {
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}}
	ngWordRepo := &fakeNGWordRepository{ngWords: []*model.NGWordModel{{Word: "ok"}, {Word: "bad"}, {Word: "never checked"}}, hitWords: []string{"bad"}}
	logger := &fakeLogger{}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, &fakeLivecommentReportRepository{}, ngWordRepo, logger)

	_, err := u.Create(context.Background(), 1, 10, "this is bad", 0)
	if !errors.Is(err, ErrSpamLivecomment) {
		t.Fatalf("err = %v, want ErrSpamLivecomment", err)
	}
	// 当たった NG ワードまでログを出し、そこで判定を打ち切る
	if want := []string{"[hitSpam=0] comment = this is bad", "[hitSpam=1] comment = this is bad"}; !slices.Equal(logger.lines, want) {
		t.Errorf("log lines = %q, want %q", logger.lines, want)
	}
}

func TestLivecommentUsecase_Report(t *testing.T) {
	want := &model.LivecommentReport{ID: 7}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}}
	livecommentRepo := &fakeLivecommentRepository{livecommentModel: &model.LivecommentModel{ID: 50, LivestreamID: 10}}
	reportRepo := &fakeLivecommentReportRepository{createID: 7, report: want}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, reportRepo, &fakeNGWordRepository{}, &fakeLogger{}).(*livecommentUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Report(context.Background(), 3, 10, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if livestreamRepo.gotID != 10 {
		t.Errorf("livestream id = %d, want 10", livestreamRepo.gotID)
	}
	if want := []string{"FindByID"}; !slices.Equal(livecommentRepo.calls, want) || livecommentRepo.gotID != 50 {
		t.Errorf("livecomment calls = %v, id = %d", livecommentRepo.calls, livecommentRepo.gotID)
	}
	wantCreated := model.LivecommentReportModel{UserID: 3, LivestreamID: 10, LivecommentID: 50, CreatedAt: 1700000000}
	if *reportRepo.gotCreated != wantCreated {
		t.Errorf("created = %+v, want %+v", *reportRepo.gotCreated, wantCreated)
	}
	if want := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(reportRepo.calls, want) {
		t.Errorf("report calls = %v, want %v", reportRepo.calls, want)
	}
	if reportRepo.gotID != 7 {
		t.Errorf("re-read id = %d, want 7", reportRepo.gotID)
	}
}

func TestLivecommentUsecase_Report_Errors(t *testing.T) {
	boom := errors.New("boom")
	livestream := &model.LivestreamModel{ID: 10, UserID: 2}
	livecomment := &model.LivecommentModel{ID: 50, LivestreamID: 10}

	tests := []struct {
		name            string
		livestreamRepo  *fakeLivestreamRepository
		livecommentRepo *fakeLivecommentRepository
		reportRepo      *fakeLivecommentReportRepository
		wantErr         error
		wantReportCalls []string
	}{
		{
			name:            "livestream not found",
			livestreamRepo:  &fakeLivestreamRepository{err: repository.ErrNotFound},
			livecommentRepo: &fakeLivecommentRepository{},
			reportRepo:      &fakeLivecommentReportRepository{},
			wantErr:         ErrLivestreamNotFound,
		},
		{
			name:            "livestream repository error",
			livestreamRepo:  &fakeLivestreamRepository{err: boom},
			livecommentRepo: &fakeLivecommentRepository{},
			reportRepo:      &fakeLivecommentReportRepository{},
			wantErr:         boom,
		},
		{
			name:            "livecomment not found",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: livestream},
			livecommentRepo: &fakeLivecommentRepository{findErr: repository.ErrNotFound},
			reportRepo:      &fakeLivecommentReportRepository{},
			wantErr:         ErrLivecommentNotFound,
		},
		{
			name:            "livecomment repository error",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: livestream},
			livecommentRepo: &fakeLivecommentRepository{findErr: boom},
			reportRepo:      &fakeLivecommentReportRepository{},
			wantErr:         boom,
		},
		{
			name:            "create fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: livestream},
			livecommentRepo: &fakeLivecommentRepository{livecommentModel: livecomment},
			reportRepo:      &fakeLivecommentReportRepository{createErr: boom},
			wantErr:         boom,
			wantReportCalls: []string{"Create"},
		},
		{
			name:            "fill fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: livestream},
			livecommentRepo: &fakeLivecommentRepository{livecommentModel: livecomment},
			reportRepo:      &fakeLivecommentReportRepository{createID: 7, err: boom},
			wantErr:         boom,
			wantReportCalls: []string{"Create", "FindWithDetailsByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, tt.livecommentRepo, tt.reportRepo, &fakeNGWordRepository{}, &fakeLogger{})
			_, err := u.Report(context.Background(), 3, 10, 50)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			// ライブ配信やライブコメントが無い場合は報告を登録しない
			if !slices.Equal(tt.reportRepo.calls, tt.wantReportCalls) {
				t.Errorf("report calls = %v, want %v", tt.reportRepo.calls, tt.wantReportCalls)
			}
		})
	}
}
