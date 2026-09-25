package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

func TestNGWordUsecase_FindAllByLivestreamID(t *testing.T) {
	want := []*model.NGWordModel{{ID: 1, Word: "bad"}}
	repo := &fakeNGWordRepository{ngWords: want}
	u := NewNGWordUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, &fakeLivecommentRepository{}, repo)

	got, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if repo.gotUserID != 1 || repo.gotLivestreamID != 10 {
		t.Errorf("userID = %d, livestreamID = %d", repo.gotUserID, repo.gotLivestreamID)
	}
}

func TestNGWordUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewNGWordUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, &fakeLivecommentRepository{}, &fakeNGWordRepository{err: boom})

	_, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestNGWordUsecase_Moderate(t *testing.T) {
	livestreamRepo := &fakeLivestreamRepository{livestreamModels: []*model.LivestreamModel{{ID: 10, UserID: 2}}}
	livecommentRepo := &fakeLivecommentRepository{}
	// 登録したユーザによらず、ライブ配信の全 NG ワードで削除する
	ngWordRepo := &fakeNGWordRepository{createID: 7, ngWords: []*model.NGWordModel{{Word: "new"}, {Word: "old"}}}
	u := NewNGWordUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, ngWordRepo).(*ngWordUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	wordID, err := u.Moderate(context.Background(), 2, 10, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wordID != 7 {
		t.Errorf("wordID = %d, want 7", wordID)
	}
	// 配信者自身のライブ配信かを、ライブ配信 ID とセッションのユーザ ID で検索して確認する
	if livestreamRepo.gotID != 10 || livestreamRepo.gotUserID != 2 {
		t.Errorf("livestream searched by id = %d, userID = %d", livestreamRepo.gotID, livestreamRepo.gotUserID)
	}
	wantCreated := model.NGWordModel{UserID: 2, LivestreamID: 10, Word: "new", CreatedAt: 1700000000}
	if *ngWordRepo.gotCreated != wantCreated {
		t.Errorf("created = %+v, want %+v", *ngWordRepo.gotCreated, wantCreated)
	}
	// 登録してから、ライブ配信の NG ワードを取り直す
	if want := []string{"Create", "FindAllByLivestreamID"}; !slices.Equal(ngWordRepo.calls, want) {
		t.Errorf("NG word calls = %v, want %v", ngWordRepo.calls, want)
	}
	if want := []string{"new", "old"}; !slices.Equal(livecommentRepo.deletedWords, want) {
		t.Errorf("deleted words = %v, want %v", livecommentRepo.deletedWords, want)
	}
	if livecommentRepo.gotLivestreamID != 10 {
		t.Errorf("livestreamID = %d, want 10", livecommentRepo.gotLivestreamID)
	}
}

func TestNGWordUsecase_Moderate_Errors(t *testing.T) {
	boom := errors.New("boom")
	owned := []*model.LivestreamModel{{ID: 10, UserID: 2}}

	tests := []struct {
		name            string
		livestreamRepo  *fakeLivestreamRepository
		livecommentRepo *fakeLivecommentRepository
		ngWordRepo      *fakeNGWordRepository
		wantErr         error
		wantNGWordCalls []string
	}{
		{
			// 他の配信者のライブ配信・存在しないライブ配信のどちらも、検索結果が空になる
			name:            "not the owner",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModels: nil},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{},
			wantErr:         ErrNotLivestreamOwner,
		},
		{
			name:            "livestream repository error",
			livestreamRepo:  &fakeLivestreamRepository{err: boom},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{},
			wantErr:         boom,
		},
		{
			name:            "create fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModels: owned},
			livecommentRepo: &fakeLivecommentRepository{},
			ngWordRepo:      &fakeNGWordRepository{createErr: boom},
			wantErr:         boom,
			wantNGWordCalls: []string{"Create"},
		},
		{
			name:            "delete fails",
			livestreamRepo:  &fakeLivestreamRepository{livestreamModels: owned},
			livecommentRepo: &fakeLivecommentRepository{deleteErr: boom},
			ngWordRepo:      &fakeNGWordRepository{ngWords: []*model.NGWordModel{{Word: "bad"}}},
			wantErr:         boom,
			wantNGWordCalls: []string{"Create", "FindAllByLivestreamID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewNGWordUsecase(&fakeTxManager{}, tt.livestreamRepo, tt.livecommentRepo, tt.ngWordRepo)
			_, err := u.Moderate(context.Background(), 2, 10, "bad")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !slices.Equal(tt.ngWordRepo.calls, tt.wantNGWordCalls) {
				t.Errorf("NG word calls = %v, want %v", tt.ngWordRepo.calls, tt.wantNGWordCalls)
			}
		})
	}
}
