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

// newReactionRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、models (取得に失敗させる場合は err) を返す fakeReactionRepository を返す。
// ライブ配信 10 のリアクションを取得すること、limit 付きの場合はその値を gotLimit に取り出す。
func newReactionRepositoryForFindAll(t *testing.T, calls *[]string, gotLimit *domain.Limit, models []*domain.ReactionModel, err error) *fakeReactionRepository {
	checkLivestreamID := func(livestreamID domain.LivestreamID) {
		if livestreamID != 10 {
			t.Errorf("livestreamID = %d, want 10", livestreamID)
		}
	}
	return &fakeReactionRepository{
		findAllByLivestreamIDOrderByCreatedAtDesc: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID) ([]*domain.ReactionModel, error) {
			*calls = append(*calls, "FindAllByLivestreamIDOrderByCreatedAtDesc")
			checkLivestreamID(livestreamID)
			return models, err
		},
		findAllByLivestreamIDOrderByCreatedAtDescLimited: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.ReactionModel, error) {
			*calls = append(*calls, "FindAllByLivestreamIDOrderByCreatedAtDescLimited")
			checkLivestreamID(livestreamID)
			*gotLimit = limit
			return models, err
		},
	}
}

func TestReactionUsecase_FindAllByLivestreamID(t *testing.T) {
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
			f := testLivestreamFixture()
			var calls []string
			var gotLimit domain.Limit
			reactionRepo := newReactionRepositoryForFindAll(t, &calls, &gotLimit, []*domain.ReactionModel{testReactionModel1, testReactionModel2}, nil)
			u := NewReactionUsecase(&fakeTxManager{}, reactionRepo, f.reactionFiller(testLivestreamModel1))

			got, err := u.FindAllByLivestreamID(context.Background(), 10, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// repository が返した順序のまま
			want := []*domain.ReactionDetail{f.reaction(testReactionModel1, testLivestreamModel1), f.reaction(testReactionModel2, testLivestreamModel1)}
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

func TestReactionUsecase_FindAllByLivestreamID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		models  []*domain.ReactionModel
		err     error
		wantErr error
		wantMsg string
	}{
		{name: "get reactions fails", err: boom, wantErr: boom, wantMsg: "failed to get reactions: boom"},
		{
			// ライブ配信 2 は引けないので組み立てに失敗する
			name:    "fill fails",
			models:  []*domain.ReactionModel{testReactionModel3},
			wantErr: repository.ErrNotFound,
			wantMsg: "failed to get reactions: failed to get livestream of reaction 3: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var gotLimit domain.Limit
			reactionRepo := newReactionRepositoryForFindAll(t, &calls, &gotLimit, tt.models, tt.err)
			u := NewReactionUsecase(&fakeTxManager{}, reactionRepo, testLivestreamFixture().reactionFiller(testLivestreamModel1))

			_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

// newReactionRepositoryForCreate は Create で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeReactionRepository を返す。
// 登録したリアクションは created に取り出し、ID 100 で読み直したリアクションとして reaction を返す。
func newReactionRepositoryForCreate(t *testing.T, calls *[]string, created **domain.ReactionModel, reaction *domain.ReactionModel, createErr, findErr error) *fakeReactionRepository {
	return &fakeReactionRepository{
		create: func(_ context.Context, _ repository.Querier, r *domain.ReactionModel) (domain.ReactionID, error) {
			*calls = append(*calls, "Create")
			*created = r
			return 100, createErr
		},
		findByID: func(_ context.Context, _ repository.Querier, id domain.ReactionID) (*domain.ReactionModel, error) {
			*calls = append(*calls, "FindByID")
			if id != 100 {
				t.Errorf("id = %d, want 100", id)
			}
			return reaction, findErr
		},
	}
}

// testCreatedReactionModel は登録したリアクション (ID 100) を読み直した結果。
var testCreatedReactionModel = &domain.ReactionModel{ID: 100, UserID: 43, LivestreamID: 1, EmojiName: "tada", CreatedAt: 1700000000}

func TestReactionUsecase_Create(t *testing.T) {
	f := testLivestreamFixture()
	var calls []string
	var created *domain.ReactionModel
	reactionRepo := newReactionRepositoryForCreate(t, &calls, &created, testCreatedReactionModel, nil, nil)
	u := NewReactionUsecase(&fakeTxManager{}, reactionRepo, f.reactionFiller(testLivestreamModel1)).(*reactionUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 43, 1, "tada")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := f.reaction(testCreatedReactionModel, testLivestreamModel1); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// 登録してから、登録した ID で読み直す
	if wantCalls := []string{"Create", "FindByID"}; !slices.Equal(calls, wantCalls) {
		t.Errorf("calls = %v, want %v", calls, wantCalls)
	}
	wantCreated := domain.ReactionModel{UserID: 43, LivestreamID: 1, EmojiName: "tada", CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
}

func TestReactionUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		createErr error
		findErr   error
		// modify はデータを欠けさせる
		modify    func(f *livestreamFixture)
		wantErr   error
		wantMsg   string
		wantCalls []string
	}{
		{name: "create fails", createErr: boom, wantErr: boom, wantMsg: "failed to insert reaction: boom", wantCalls: []string{"Create"}},
		{name: "re-read fails", findErr: boom, wantErr: boom, wantMsg: "failed to fill reaction: boom", wantCalls: []string{"Create", "FindByID"}},
		{
			name:      "fill fails",
			modify:    func(f *livestreamFixture) { delete(f.users, 43) },
			wantErr:   repository.ErrNotFound,
			wantMsg:   "failed to fill reaction: failed to get user of reaction 100: not found",
			wantCalls: []string{"Create", "FindByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			if tt.modify != nil {
				tt.modify(f)
			}
			var calls []string
			var created *domain.ReactionModel
			reactionRepo := newReactionRepositoryForCreate(t, &calls, &created, testCreatedReactionModel, tt.createErr, tt.findErr)
			u := NewReactionUsecase(&fakeTxManager{}, reactionRepo, f.reactionFiller(testLivestreamModel1))
			_, err := u.Create(context.Background(), 43, 1, "tada")
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", calls, tt.wantCalls)
			}
		})
	}
}
