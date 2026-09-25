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

// newLivecommentRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、livecomments (取得に失敗させる場合は err) を返す fakeLivecommentRepository を返す。
// ライブ配信 10 のライブコメントを取得すること、limit 付きの場合はその値を gotLimit に取り出す。
func newLivecommentRepositoryForFindAll(t *testing.T, calls *[]string, gotLimit *model.Limit, livecomments []*model.Livecomment, err error) *fakeLivecommentRepository {
	checkLivestreamID := func(livestreamID model.LivestreamID) {
		if livestreamID != 10 {
			t.Errorf("livestreamID = %d, want 10", livestreamID)
		}
	}
	return &fakeLivecommentRepository{
		findAllWithDetailsByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error) {
			*calls = append(*calls, "FindAllWithDetailsByLivestreamID")
			checkLivestreamID(livestreamID)
			return livecomments, err
		},
		findAllWithDetailsByLivestreamIDLimited: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID, limit model.Limit) ([]*model.Livecomment, error) {
			*calls = append(*calls, "FindAllWithDetailsByLivestreamIDLimited")
			checkLivestreamID(livestreamID)
			*gotLimit = limit
			return livecomments, err
		},
	}
}

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
			var calls []string
			var gotLimit model.Limit
			repo := newLivecommentRepositoryForFindAll(t, &calls, &gotLimit, want, nil)
			u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeLivecommentReportRepository{}, &fakeNGWordRepository{}, &fakeLogger{})

			got, err := u.FindAllByLivestreamID(context.Background(), 10, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", calls, tt.wantCalls)
			}
			if gotLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", gotLimit, tt.wantLimit)
			}
		})
	}
}

func TestLivecommentUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	var calls []string
	var gotLimit model.Limit
	repo := newLivecommentRepositoryForFindAll(t, &calls, &gotLimit, nil, boom)
	u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeLivecommentReportRepository{}, &fakeNGWordRepository{}, &fakeLogger{})

	_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivecommentUsecase_FindAllReportsByLivestreamID(t *testing.T) {
	want := []*model.LivecommentReport{{ID: 1}}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 1}}
	reportRepo := &fakeLivecommentReportRepository{
		findAllWithDetailsByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
			if livestreamID != 10 {
				t.Errorf("report livestream id = %d, want 10", livestreamID)
			}
			return want, nil
		},
	}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, reportRepo, &fakeNGWordRepository{}, &fakeLogger{})

	got, err := u.FindAllReportsByLivestreamID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if livestreamRepo.gotID != 10 {
		t.Errorf("livestream id = %d, want 10", livestreamRepo.gotID)
	}
}

func TestLivecommentUsecase_FindAllReportsByLivestreamID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name           string
		livestreamRepo *fakeLivestreamRepository
		// reportsErr は報告の取得が返すエラー
		reportsErr      error
		check           func(t *testing.T, err error)
		wantReportCalls int
	}{
		{
			name:           "not the owner",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}},
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
			check: func(t *testing.T, err error) {
				if err == nil || errors.Is(err, ErrLivestreamNotFound) || errors.Is(err, ErrNotLivestreamOwner) {
					t.Errorf("err = %v, want a plain error", err)
				}
			},
		},
		{
			name:            "report repository error",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 1}},
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
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, &fakeLivecommentRepository{}, reportRepo, &fakeNGWordRepository{}, &fakeLogger{})
			_, err := u.FindAllReportsByLivestreamID(context.Background(), 1, 10)
			tt.check(t, err)
			if reportCalls != tt.wantReportCalls {
				t.Errorf("report calls = %d, want %d", reportCalls, tt.wantReportCalls)
			}
		})
	}
}

// newNGWordRepositoryForCreate はスパム判定用の fakeNGWordRepository を返す。
// 配信者 (userID 2) のライブ配信 10 の NG ワードとして ngWords (取得に失敗させる場合は findErr) を返し、
// Matches では hitWords に含まれる NG ワードを当たりとする (matchErr を設定した場合はエラー)。判定した NG ワードは matched に記録する。
func newNGWordRepositoryForCreate(t *testing.T, ngWords []*model.NGWordModel, findErr error, hitWords []string, matchErr error, matched *[]string) *fakeNGWordRepository {
	return &fakeNGWordRepository{
		findAllByUserIDAndLivestreamID: func(_ context.Context, _ repository.Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
			// NG ワードは投稿者ではなく配信者のものを使う
			if userID != 2 || livestreamID != 10 {
				t.Errorf("NG words of userID = %d, livestreamID = %d, want 2, 10", userID, livestreamID)
			}
			return ngWords, findErr
		},
		matches: func(_ context.Context, _ repository.Querier, comment string, word string) (bool, error) {
			*matched = append(*matched, word)
			if matchErr != nil {
				return false, matchErr
			}
			return slices.Contains(hitWords, word), nil
		},
	}
}

// newLivecommentRepositoryForCreate は Create で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeLivecommentRepository を返す。
// 登録したライブコメントは created に取り出し、ID 100 で読み直したライブコメントとして livecomment を返す。
func newLivecommentRepositoryForCreate(t *testing.T, calls *[]string, created **model.LivecommentModel, livecomment *model.Livecomment, createErr, fillErr error) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		create: func(_ context.Context, _ repository.Querier, l *model.LivecommentModel) (model.LivecommentID, error) {
			*calls = append(*calls, "Create")
			*created = l
			return 100, createErr
		},
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.LivecommentID) (*model.Livecomment, error) {
			*calls = append(*calls, "FindWithDetailsByID")
			if id != 100 {
				t.Errorf("re-read id = %d, want 100", id)
			}
			return livecomment, fillErr
		},
	}
}

func TestLivecommentUsecase_Create(t *testing.T) {
	want := &model.Livecomment{ID: 100, Comment: "hello"}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}}
	var livecommentCalls []string
	var created *model.LivecommentModel
	livecommentRepo := newLivecommentRepositoryForCreate(t, &livecommentCalls, &created, want, nil, nil)
	var matched []string
	ngWordRepo := newNGWordRepositoryForCreate(t, []*model.NGWordModel{{Word: "bad"}, {Word: "evil"}}, nil, nil, nil, &matched)
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
	if want := []string{"bad", "evil"}; !slices.Equal(matched, want) {
		t.Errorf("matched words = %v, want %v", matched, want)
	}
	// NG ワードごとに判定結果をログに出す (移行前と同じ形式)
	if want := []string{"[hitSpam=0] comment = hello", "[hitSpam=0] comment = hello"}; !slices.Equal(logger.lines, want) {
		t.Errorf("log lines = %q, want %q", logger.lines, want)
	}
	if want := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(livecommentCalls, want) {
		t.Errorf("calls = %v, want %v", livecommentCalls, want)
	}
	wantCreated := model.LivecommentModel{UserID: 1, LivestreamID: 10, Comment: "hello", Tip: 500, CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
}

func TestLivecommentUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")
	owned := &model.LivestreamModel{ID: 10, UserID: 2}

	tests := []struct {
		name           string
		livestreamRepo *fakeLivestreamRepository
		// livecommentCreateErr, livecommentFillErr はライブコメントの登録・取り直しが返すエラー
		livecommentCreateErr error
		livecommentFillErr   error
		// ngWords, ngWordsErr, hitWords, matchErr はスパム判定の NG ワードの取得・判定の結果
		ngWords    []*model.NGWordModel
		ngWordsErr error
		hitWords   []string
		matchErr   error
		wantErr    error
		wantCalls  []string
	}{
		{
			name:           "livestream not found",
			livestreamRepo: &fakeLivestreamRepository{err: repository.ErrNotFound},
			wantErr:        ErrLivestreamNotFound,
		},
		{
			name:           "spam",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: owned},
			ngWords:        []*model.NGWordModel{{Word: "ok"}, {Word: "bad"}},
			hitWords:       []string{"bad"},
			wantErr:        ErrSpamLivecomment,
		},
		{
			name:           "NG words repository error",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: owned},
			ngWordsErr:     boom,
			wantErr:        boom,
		},
		{
			name:           "match error",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: owned},
			ngWords:        []*model.NGWordModel{{Word: "bad"}},
			matchErr:       boom,
			wantErr:        boom,
		},
		{
			name:                 "create fails",
			livestreamRepo:       &fakeLivestreamRepository{livestreamModel: owned},
			livecommentCreateErr: boom,
			wantErr:              boom,
			wantCalls:            []string{"Create"},
		},
		{
			name:               "fill fails",
			livestreamRepo:     &fakeLivestreamRepository{livestreamModel: owned},
			livecommentFillErr: boom,
			wantErr:            boom,
			wantCalls:          []string{"Create", "FindWithDetailsByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched []string
			ngWordRepo := newNGWordRepositoryForCreate(t, tt.ngWords, tt.ngWordsErr, tt.hitWords, tt.matchErr, &matched)
			var livecommentCalls []string
			var created *model.LivecommentModel
			livecommentRepo := newLivecommentRepositoryForCreate(t, &livecommentCalls, &created, nil, tt.livecommentCreateErr, tt.livecommentFillErr)
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, livecommentRepo, &fakeLivecommentReportRepository{}, ngWordRepo, &fakeLogger{})
			_, err := u.Create(context.Background(), 1, 10, "this is bad", 0)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			// スパム判定などで失敗した場合はライブコメントを登録しない
			if !slices.Equal(livecommentCalls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", livecommentCalls, tt.wantCalls)
			}
		})
	}
}

func TestLivecommentUsecase_Create_LogsHitSpam(t *testing.T) {
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}}
	var matched []string
	ngWordRepo := newNGWordRepositoryForCreate(t, []*model.NGWordModel{{Word: "ok"}, {Word: "bad"}, {Word: "never checked"}}, nil, []string{"bad"}, nil, &matched)
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
	if want := []string{"ok", "bad"}; !slices.Equal(matched, want) {
		t.Errorf("matched words = %v, want %v", matched, want)
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

func TestLivecommentUsecase_Report(t *testing.T) {
	want := &model.LivecommentReport{ID: 7}
	livestreamRepo := &fakeLivestreamRepository{livestreamModel: &model.LivestreamModel{ID: 10, UserID: 2}}
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

func TestLivecommentUsecase_Report_Errors(t *testing.T) {
	boom := errors.New("boom")
	livestream := &model.LivestreamModel{ID: 10, UserID: 2}
	livecomment := &model.LivecommentModel{ID: 50, LivestreamID: 10}

	tests := []struct {
		name           string
		livestreamRepo *fakeLivestreamRepository
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
			name:           "livestream not found",
			livestreamRepo: &fakeLivestreamRepository{err: repository.ErrNotFound},
			wantErr:        ErrLivestreamNotFound,
		},
		{
			name:           "livestream repository error",
			livestreamRepo: &fakeLivestreamRepository{err: boom},
			wantErr:        boom,
		},
		{
			name:           "livecomment not found",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: livestream},
			livecommentErr: repository.ErrNotFound,
			wantErr:        ErrLivecommentNotFound,
		},
		{
			name:           "livecomment repository error",
			livestreamRepo: &fakeLivestreamRepository{livestreamModel: livestream},
			livecommentErr: boom,
			wantErr:        boom,
		},
		{
			name:            "create fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: livestream},
			livecomment:     livecomment,
			reportCreateErr: boom,
			wantErr:         boom,
			wantReportCalls: []string{"Create"},
		},
		{
			name:            "fill fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModel: livestream},
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
			u := NewLivecommentUsecase(&fakeTxManager{}, tt.livestreamRepo, livecommentRepo, reportRepo, &fakeNGWordRepository{}, &fakeLogger{})
			_, err := u.Report(context.Background(), 3, 10, 50)
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
