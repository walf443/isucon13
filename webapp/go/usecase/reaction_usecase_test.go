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

// newReactionRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、reactions (取得に失敗させる場合は err) を返す fakeReactionRepository を返す。
// ライブ配信 10 のリアクションを取得すること、limit 付きの場合はその値を gotLimit に取り出す。
func newReactionRepositoryForFindAll(t *testing.T, calls *[]string, gotLimit *domain.Limit, reactions []*domain.Reaction, err error) *fakeReactionRepository {
	checkLivestreamID := func(livestreamID domain.LivestreamID) {
		if livestreamID != 10 {
			t.Errorf("livestreamID = %d, want 10", livestreamID)
		}
	}
	return &fakeReactionRepository{
		findAllWithDetailsByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Reaction, error) {
			*calls = append(*calls, "FindAllWithDetailsByLivestreamID")
			checkLivestreamID(livestreamID)
			return reactions, err
		},
		findAllWithDetailsByLivestreamIDLimited: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Reaction, error) {
			*calls = append(*calls, "FindAllWithDetailsByLivestreamIDLimited")
			checkLivestreamID(livestreamID)
			*gotLimit = limit
			return reactions, err
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
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetailsByLivestreamID"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsByLivestreamIDLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*domain.Reaction{{ID: 1}}
			var calls []string
			var gotLimit domain.Limit
			u := NewReactionUsecase(&fakeTxManager{}, newReactionRepositoryForFindAll(t, &calls, &gotLimit, want, nil))

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

func TestReactionUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	var calls []string
	var gotLimit domain.Limit
	u := NewReactionUsecase(&fakeTxManager{}, newReactionRepositoryForFindAll(t, &calls, &gotLimit, nil, boom))

	_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

// newReactionRepositoryForCreate は Create で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeReactionRepository を返す。
// 登録したリアクションは created に取り出し、ID 100 で読み直したリアクションとして reaction を返す。
func newReactionRepositoryForCreate(t *testing.T, calls *[]string, created **domain.ReactionModel, reaction *domain.Reaction, createErr, fillErr error) *fakeReactionRepository {
	return &fakeReactionRepository{
		create: func(_ context.Context, _ repository.Querier, r *domain.ReactionModel) (domain.ReactionID, error) {
			*calls = append(*calls, "Create")
			*created = r
			return 100, createErr
		},
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id domain.ReactionID) (*domain.Reaction, error) {
			*calls = append(*calls, "FindWithDetailsByID")
			if id != 100 {
				t.Errorf("id = %d, want 100", id)
			}
			return reaction, fillErr
		},
	}
}

func TestReactionUsecase_Create(t *testing.T) {
	want := &domain.Reaction{ID: 100, EmojiName: "tada"}
	var calls []string
	var created *domain.ReactionModel
	u := NewReactionUsecase(&fakeTxManager{}, newReactionRepositoryForCreate(t, &calls, &created, want, nil, nil)).(*reactionUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 1, 10, "tada")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// 登録してから、登録した ID で読み直す
	if wantCalls := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(calls, wantCalls) {
		t.Errorf("calls = %v, want %v", calls, wantCalls)
	}
	wantCreated := domain.ReactionModel{UserID: 1, LivestreamID: 10, EmojiName: "tada", CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
}

func TestReactionUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		createErr error
		fillErr   error
		wantCalls []string
	}{
		{name: "create fails", createErr: boom, wantCalls: []string{"Create"}},
		{name: "fill fails", fillErr: boom, wantCalls: []string{"Create", "FindWithDetailsByID"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var created *domain.ReactionModel
			u := NewReactionUsecase(&fakeTxManager{}, newReactionRepositoryForCreate(t, &calls, &created, nil, tt.createErr, tt.fillErr))
			_, err := u.Create(context.Background(), 1, 10, "tada")
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", calls, tt.wantCalls)
			}
		})
	}
}
