package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

func TestReactionUsecase_FindAllByLivestreamID(t *testing.T) {
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
			want := []*model.Reaction{{ID: 1}}
			repo := &fakeReactionRepository{reactions: want}
			u := NewReactionUsecase(&fakeTxManager{}, repo)

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

func TestReactionUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewReactionUsecase(&fakeTxManager{}, &fakeReactionRepository{err: boom})

	_, err := u.FindAllByLivestreamID(context.Background(), 10, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestReactionUsecase_Create(t *testing.T) {
	want := &model.Reaction{ID: 100, EmojiName: "tada"}
	repo := &fakeReactionRepository{createID: 100, reaction: want}
	u := NewReactionUsecase(&fakeTxManager{}, repo).(*reactionUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 1, 10, "tada")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// 登録してから、登録した ID で読み直す
	if wantCalls := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(repo.calls, wantCalls) {
		t.Errorf("calls = %v, want %v", repo.calls, wantCalls)
	}
	wantCreated := model.ReactionModel{UserID: 1, LivestreamID: 10, EmojiName: "tada", CreatedAt: 1700000000}
	if *repo.gotCreated != wantCreated {
		t.Errorf("created = %+v, want %+v", *repo.gotCreated, wantCreated)
	}
	if repo.gotID != 100 {
		t.Errorf("id = %d, want 100", repo.gotID)
	}
}

func TestReactionUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		repo      *fakeReactionRepository
		wantCalls []string
	}{
		{name: "create fails", repo: &fakeReactionRepository{createErr: boom}, wantCalls: []string{"Create"}},
		{name: "fill fails", repo: &fakeReactionRepository{createID: 100, err: boom}, wantCalls: []string{"Create", "FindWithDetailsByID"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewReactionUsecase(&fakeTxManager{}, tt.repo)
			_, err := u.Create(context.Background(), 1, 10, "tada")
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
			if !slices.Equal(tt.repo.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", tt.repo.calls, tt.wantCalls)
			}
		})
	}
}
