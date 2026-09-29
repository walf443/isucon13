package usecase

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newLivecommentRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、models (取得に失敗させる場合は err) を返す fakeLivecommentRepository を返す。
// ライブ配信 10 のライブコメントを取得すること、limit 付きの場合はその値を gotLimit に取り出す。
func newLivecommentRepositoryForFindAll(t *testing.T, calls *[]string, gotLimit *domain.Limit, models []*domain.Livecomment, err error) *fakeLivecommentRepository {
	checkLivestreamID := func(livestreamID domain.LivestreamID) {
		if livestreamID != 10 {
			t.Errorf("livestreamID = %d, want 10", livestreamID)
		}
	}
	return &fakeLivecommentRepository{
		findAllByLivestreamIDOrderByCreatedAtDesc: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Livecomment, error) {
			*calls = append(*calls, "FindAllByLivestreamIDOrderByCreatedAtDesc")
			checkLivestreamID(livestreamID)
			return models, err
		},
		findAllByLivestreamIDOrderByCreatedAtDescLimited: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Livecomment, error) {
			*calls = append(*calls, "FindAllByLivestreamIDOrderByCreatedAtDescLimited")
			checkLivestreamID(livestreamID)
			*gotLimit = limit
			return models, err
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
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllByLivestreamIDOrderByCreatedAtDesc"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllByLivestreamIDOrderByCreatedAtDescLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testDetailFixture()
			var calls []string
			var gotLimit domain.Limit
			repo := newLivecommentRepositoryForFindAll(t, &calls, &gotLimit, []*domain.Livecomment{testLivecomment1, testLivecomment2}, nil)
			u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeNGWordRepository{}, f.livecommentFiller(testLivestream1), &fakeLogger{})

			got, err := u.FindAllByLivestreamID(context.Background(), 10, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// repository が返した順序のまま
			want := []*domain.LivecommentDetail{f.livecomment(testLivecomment1, testLivestream1), f.livecomment(testLivecomment2, testLivestream1)}
			if !reflect.DeepEqual(got, want) {
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

func TestLivecommentUsecase_FindAllByLivestreamID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		models  []*domain.Livecomment
		err     error
		wantErr error
		wantMsg string
	}{
		{name: "get livecomments fails", err: boom, wantErr: boom, wantMsg: "failed to get livecomments: boom"},
		{
			// ライブ配信 2 は引けないので組み立てに失敗する
			name:    "fill fails",
			models:  []*domain.Livecomment{testLivecomment3},
			wantErr: errMissingDetail,
			wantMsg: "failed to get livecomments: failed to get livestream of livecomment 52: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var gotLimit domain.Limit
			repo := newLivecommentRepositoryForFindAll(t, &calls, &gotLimit, tt.models, tt.err)
			u := NewLivecommentUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, repo, &fakeNGWordRepository{}, testDetailFixture().livecommentFiller(testLivestream1), &fakeLogger{})

			_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

// newNGWordRepositoryForCreate はスパム判定用の fakeNGWordRepository を返す。
// 配信者 (userID 2) のライブ配信 10 の NG ワードとして ngWords (取得に失敗させる場合は findErr) を返し、
// Matches では hitWords に含まれる NG ワードを当たりとする (matchErr を設定した場合はエラー)。判定した NG ワードは matched に記録する。
func newNGWordRepositoryForCreate(t *testing.T, ngWords []*domain.NGWord, findErr error, hitWords []string, matchErr error, matched *[]string) *fakeNGWordRepository {
	return &fakeNGWordRepository{
		findAllByUserIDAndLivestreamID: func(_ context.Context, _ repository.Querier, userID domain.UserID, livestreamID domain.LivestreamID) ([]*domain.NGWord, error) {
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
// 登録したライブコメントは created に取り出し、ID 100 で読み直したライブコメントとして testCreatedLivecomment を返す。
func newLivecommentRepositoryForCreate(t *testing.T, calls *[]string, created **domain.Livecomment, createErr, findErr error) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		create: func(_ context.Context, _ repository.Querier, l *domain.Livecomment) (domain.LivecommentID, error) {
			*calls = append(*calls, "Create")
			*created = l
			return 100, createErr
		},
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentID) (*domain.Livecomment, error) {
			*calls = append(*calls, "FindByID")
			if id != 100 {
				t.Errorf("re-read id = %d, want 100", id)
			}
			return testCreatedLivecomment, findErr
		},
	}
}

var (
	// testOwnedLivestream は投稿先のライブ配信 10 (配信者は ID 2 のユーザ)。
	testOwnedLivestream = &domain.Livestream{ID: 10, UserID: 2, Title: "owned"}
	// testCreatedLivecomment は投稿したライブコメント (ID 100) を読み直した結果。
	testCreatedLivecomment = &domain.Livecomment{ID: 100, UserID: 43, LivestreamID: 10, Comment: "hello", Tip: 500, CreatedAt: 1700000000}
)

// livecommentFixtureForCreate は投稿先のライブ配信 10 の配信者 (ID 2) を加えたテスト用のデータを返す。
func livecommentFixtureForCreate() *detailFixture {
	f := testDetailFixture()
	f.users[2] = &domain.User{ID: 2, Name: "owner"}
	return f
}

func TestLivecommentUsecase_Create(t *testing.T) {
	f := livecommentFixtureForCreate()
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 10, testOwnedLivestream, nil)
	var livecommentCalls []string
	var created *domain.Livecomment
	livecommentRepo := newLivecommentRepositoryForCreate(t, &livecommentCalls, &created, nil, nil)
	var matched []string
	ngWordRepo := newNGWordRepositoryForCreate(t, []*domain.NGWord{{Word: "bad"}, {Word: "evil"}}, nil, nil, nil, &matched)
	logger := &fakeLogger{}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, ngWordRepo, f.livecommentFiller(testOwnedLivestream), logger).(*livecommentUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 43, 10, "hello", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := f.livecomment(testCreatedLivecomment, testOwnedLivestream); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if want := []string{"bad", "evil"}; !slices.Equal(matched, want) {
		t.Errorf("matched words = %v, want %v", matched, want)
	}
	// NG ワードごとに判定結果をログに出す (移行前と同じ形式)
	if want := []string{"[hitSpam=0] comment = hello", "[hitSpam=0] comment = hello"}; !slices.Equal(logger.lines, want) {
		t.Errorf("log lines = %q, want %q", logger.lines, want)
	}
	if want := []string{"Create", "FindByID"}; !slices.Equal(livecommentCalls, want) {
		t.Errorf("calls = %v, want %v", livecommentCalls, want)
	}
	wantCreated := domain.Livecomment{UserID: 43, LivestreamID: 10, Comment: "hello", Tip: 500, CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
}

func TestLivecommentUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")
	owned := &domain.Livestream{ID: 10, UserID: 2}

	tests := []struct {
		name string
		// livestream, livestreamErr はライブ配信 10 を引いた結果
		livestream    *domain.Livestream
		livestreamErr error
		// livecommentCreateErr, livecommentFindErr はライブコメントの登録・取り直しが返すエラー
		livecommentCreateErr error
		livecommentFindErr   error
		// modify はデータを欠けさせる
		modify func(f *detailFixture)
		// ngWords, ngWordsErr, hitWords, matchErr はスパム判定の NG ワードの取得・判定の結果
		ngWords    []*domain.NGWord
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
			ngWords:    []*domain.NGWord{{Word: "ok"}, {Word: "bad"}},
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
			ngWords:    []*domain.NGWord{{Word: "bad"}},
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
			name:               "re-read fails",
			livestream:         owned,
			livecommentFindErr: boom,
			wantErr:            boom,
			wantCalls:          []string{"Create", "FindByID"},
		},
		{
			name:       "fill fails",
			livestream: owned,
			modify:     func(f *detailFixture) { delete(f.users, 43) },
			wantErr:    errMissingDetail,
			wantCalls:  []string{"Create", "FindByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched []string
			ngWordRepo := newNGWordRepositoryForCreate(t, tt.ngWords, tt.ngWordsErr, tt.hitWords, tt.matchErr, &matched)
			var livecommentCalls []string
			var created *domain.Livecomment
			livecommentRepo := newLivecommentRepositoryForCreate(t, &livecommentCalls, &created, tt.livecommentCreateErr, tt.livecommentFindErr)
			f := livecommentFixtureForCreate()
			if tt.modify != nil {
				tt.modify(f)
			}
			u := NewLivecommentUsecase(&fakeTxManager{}, newLivestreamRepositoryFindingByID(t, 10, tt.livestream, tt.livestreamErr), livecommentRepo, ngWordRepo, f.livecommentFiller(testOwnedLivestream), &fakeLogger{})
			_, err := u.Create(context.Background(), 43, 10, "this is bad", 0)
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
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 10, &domain.Livestream{ID: 10, UserID: 2}, nil)
	var matched []string
	ngWordRepo := newNGWordRepositoryForCreate(t, []*domain.NGWord{{Word: "ok"}, {Word: "bad"}, {Word: "never checked"}}, nil, []string{"bad"}, nil, &matched)
	logger := &fakeLogger{}
	u := NewLivecommentUsecase(&fakeTxManager{}, livestreamRepo, &fakeLivecommentRepository{}, ngWordRepo, nil, logger)

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
