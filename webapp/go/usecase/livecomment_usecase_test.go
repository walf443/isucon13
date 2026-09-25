package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newLivecommentRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、livecomments (取得に失敗させる場合は err) を返す fakeLivecommentRepository を返す。
// ライブ配信 10 のライブコメントを取得すること、limit 付きの場合はその値を gotLimit に取り出す。
func newLivecommentRepositoryForFindAll(t *testing.T, calls *[]string, gotLimit *domain.Limit, livecomments []*domain.Livecomment, err error) *fakeLivecommentRepository {
	checkLivestreamID := func(livestreamID domain.LivestreamID) {
		if livestreamID != 10 {
			t.Errorf("livestreamID = %d, want 10", livestreamID)
		}
	}
	return &fakeLivecommentRepository{
		findAllWithDetailsByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Livecomment, error) {
			*calls = append(*calls, "FindAllWithDetailsByLivestreamID")
			checkLivestreamID(livestreamID)
			return livecomments, err
		},
		findAllWithDetailsByLivestreamIDLimited: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Livecomment, error) {
			*calls = append(*calls, "FindAllWithDetailsByLivestreamIDLimited")
			checkLivestreamID(livestreamID)
			*gotLimit = limit
			return livecomments, err
		},
	}
}

func TestLivecommentUsecase_FindAllByLivestreamID(t *testing.T) {
	limit := domain.Limit(5)

	tests := []struct {
		name      string
		limit     *domain.Limit
		wantCalls []string
		wantLimit domain.Limit
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetailsByLivestreamID"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsByLivestreamIDLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*domain.Livecomment{{ID: 1}}
			var calls []string
			var gotLimit domain.Limit
			repo := newLivecommentRepositoryForFindAll(t, &calls, &gotLimit, want, nil)
			u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeNGWordRepository{}, &fakeLogger{})

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
	var gotLimit domain.Limit
	repo := newLivecommentRepositoryForFindAll(t, &calls, &gotLimit, nil, boom)
	u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeNGWordRepository{}, &fakeLogger{})

	_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

// newNGWordRepositoryForCreate はスパム判定用の fakeNGWordRepository を返す。
// 配信者 (userID 2) のライブ配信 10 の NG ワードとして ngWords (取得に失敗させる場合は findErr) を返し、
// Matches では hitWords に含まれる NG ワードを当たりとする (matchErr を設定した場合はエラー)。判定した NG ワードは matched に記録する。
func newNGWordRepositoryForCreate(t *testing.T, ngWords []*domain.NGWordModel, findErr error, hitWords []string, matchErr error, matched *[]string) *fakeNGWordRepository {
	return &fakeNGWordRepository{
		findAllByUserIDAndLivestreamID: func(_ context.Context, _ repository.Querier, userID domain.UserID, livestreamID domain.LivestreamID) ([]*domain.NGWordModel, error) {
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
func newLivecommentRepositoryForCreate(t *testing.T, calls *[]string, created **domain.LivecommentModel, livecomment *domain.Livecomment, createErr, fillErr error) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		create: func(_ context.Context, _ repository.Querier, l *domain.LivecommentModel) (domain.LivecommentID, error) {
			*calls = append(*calls, "Create")
			*created = l
			return 100, createErr
		},
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentID) (*domain.Livecomment, error) {
			*calls = append(*calls, "FindWithDetailsByID")
			if id != 100 {
				t.Errorf("re-read id = %d, want 100", id)
			}
			return livecomment, fillErr
		},
	}
}

func TestLivecommentUsecase_Create(t *testing.T) {
	want := &domain.Livecomment{ID: 100, Comment: "hello"}
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 10, &domain.LivestreamModel{ID: 10, UserID: 2}, nil)
	var livecommentCalls []string
	var created *domain.LivecommentModel
	livecommentRepo := newLivecommentRepositoryForCreate(t, &livecommentCalls, &created, want, nil, nil)
	var matched []string
	ngWordRepo := newNGWordRepositoryForCreate(t, []*domain.NGWordModel{{Word: "bad"}, {Word: "evil"}}, nil, nil, nil, &matched)
	logger := &fakeLogger{}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, ngWordRepo, logger).(*livecommentUsecase)
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
	wantCreated := domain.LivecommentModel{UserID: 1, LivestreamID: 10, Comment: "hello", Tip: 500, CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
}

func TestLivecommentUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")
	owned := &domain.LivestreamModel{ID: 10, UserID: 2}

	tests := []struct {
		name string
		// livestream, livestreamErr はライブ配信 10 を引いた結果
		livestream    *domain.LivestreamModel
		livestreamErr error
		// livecommentCreateErr, livecommentFillErr はライブコメントの登録・取り直しが返すエラー
		livecommentCreateErr error
		livecommentFillErr   error
		// ngWords, ngWordsErr, hitWords, matchErr はスパム判定の NG ワードの取得・判定の結果
		ngWords    []*domain.NGWordModel
		ngWordsErr error
		hitWords   []string
		matchErr   error
		wantErr    error
		wantCalls  []string
	}{
		{
			name:          "livestream not found",
			livestreamErr: repository.ErrNotFound,
			wantErr:       ErrLivestreamNotFound,
		},
		{
			name:       "spam",
			livestream: owned,
			ngWords:    []*domain.NGWordModel{{Word: "ok"}, {Word: "bad"}},
			hitWords:   []string{"bad"},
			wantErr:    ErrSpamLivecomment,
		},
		{
			name:       "NG words repository error",
			livestream: owned,
			ngWordsErr: boom,
			wantErr:    boom,
		},
		{
			name:       "match error",
			livestream: owned,
			ngWords:    []*domain.NGWordModel{{Word: "bad"}},
			matchErr:   boom,
			wantErr:    boom,
		},
		{
			name:                 "create fails",
			livestream:           owned,
			livecommentCreateErr: boom,
			wantErr:              boom,
			wantCalls:            []string{"Create"},
		},
		{
			name:               "fill fails",
			livestream:         owned,
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
			var created *domain.LivecommentModel
			livecommentRepo := newLivecommentRepositoryForCreate(t, &livecommentCalls, &created, nil, tt.livecommentCreateErr, tt.livecommentFillErr)
			u := NewLivecommentUsecase(&fakeTxManager{}, newLivestreamRepositoryFindingByID(t, 10, tt.livestream, tt.livestreamErr), livecommentRepo, ngWordRepo, &fakeLogger{})
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
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 10, &domain.LivestreamModel{ID: 10, UserID: 2}, nil)
	var matched []string
	ngWordRepo := newNGWordRepositoryForCreate(t, []*domain.NGWordModel{{Word: "ok"}, {Word: "bad"}, {Word: "never checked"}}, nil, []string{"bad"}, nil, &matched)
	logger := &fakeLogger{}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, ngWordRepo, logger)

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
