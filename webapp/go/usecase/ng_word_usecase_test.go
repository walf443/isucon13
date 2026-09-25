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

func TestNGWordUsecase_FindAllByLivestreamID(t *testing.T) {
	want := []*model.NGWordModel{{ID: 1, Word: "bad"}}
	repo := &fakeNGWordRepository{
		findAllByUserIDAndLivestreamID: func(_ context.Context, _ repository.Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
			if userID != 1 || livestreamID != 10 {
				t.Errorf("userID = %d, livestreamID = %d", userID, livestreamID)
			}
			return want, nil
		},
	}
	u := NewNGWordUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, &fakeLivecommentRepository{}, repo)

	got, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestNGWordUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	repo := &fakeNGWordRepository{
		findAllByUserIDAndLivestreamID: func(context.Context, repository.Querier, model.UserID, model.LivestreamID) ([]*model.NGWordModel, error) {
			return nil, boom
		},
	}
	u := NewNGWordUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, &fakeLivecommentRepository{}, repo)

	_, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

// newLivestreamRepositoryForModerate は、ライブ配信 10 かつ配信者がユーザ 2 のライブ配信として livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
func newLivestreamRepositoryForModerate(t *testing.T, livestreams []*model.LivestreamModel, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		// 配信者自身のライブ配信かを、ライブ配信 ID とセッションのユーザ ID で検索して確認する
		findAllByIDAndUserID: func(_ context.Context, _ repository.Querier, id model.LivestreamID, userID model.UserID) ([]*model.LivestreamModel, error) {
			if id != 10 || userID != 2 {
				t.Errorf("livestream searched by id = %d, userID = %d", id, userID)
			}
			return livestreams, err
		},
	}
}

// newLivecommentRepositoryForModerate は、ライブ配信 10 のライブコメントを NG ワードで削除する fakeLivecommentRepository を返す。
// 削除に使った NG ワードは deletedWords に記録し、削除は err を返す。
func newLivecommentRepositoryForModerate(t *testing.T, deletedWords *[]string, err error) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		deleteAllByLivestreamIDMatchingNGWord: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID, word string) error {
			*deletedWords = append(*deletedWords, word)
			if livestreamID != 10 {
				t.Errorf("livestreamID = %d, want 10", livestreamID)
			}
			return err
		},
	}
}

// newNGWordRepositoryForModerate は Moderate で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeNGWordRepository を返す。
// 登録した NG ワードは created に取り出し、取り直した NG ワードとして ngWords を返す。
func newNGWordRepositoryForModerate(t *testing.T, calls *[]string, created **model.NGWordModel, ngWords []*model.NGWordModel, createErr error) *fakeNGWordRepository {
	return &fakeNGWordRepository{
		create: func(_ context.Context, _ repository.Querier, ngWord *model.NGWordModel) (model.NGWordID, error) {
			*calls = append(*calls, "Create")
			*created = ngWord
			return 7, createErr
		},
		findAllByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
			*calls = append(*calls, "FindAllByLivestreamID")
			if livestreamID != 10 {
				t.Errorf("NG words of livestreamID = %d, want 10", livestreamID)
			}
			return ngWords, nil
		},
	}
}

func TestNGWordUsecase_Moderate(t *testing.T) {
	livestreamRepo := newLivestreamRepositoryForModerate(t, []*model.LivestreamModel{{ID: 10, UserID: 2}}, nil)
	var deletedWords []string
	livecommentRepo := newLivecommentRepositoryForModerate(t, &deletedWords, nil)
	var ngWordCalls []string
	var created *model.NGWordModel
	// 登録したユーザによらず、ライブ配信の全 NG ワードで削除する
	ngWordRepo := newNGWordRepositoryForModerate(t, &ngWordCalls, &created, []*model.NGWordModel{{Word: "new"}, {Word: "old"}}, nil)
	u := NewNGWordUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, ngWordRepo).(*ngWordUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	wordID, err := u.Moderate(context.Background(), 2, 10, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wordID != 7 {
		t.Errorf("wordID = %d, want 7", wordID)
	}
	wantCreated := model.NGWordModel{UserID: 2, LivestreamID: 10, Word: "new", CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
	// 登録してから、ライブ配信の NG ワードを取り直す
	if want := []string{"Create", "FindAllByLivestreamID"}; !slices.Equal(ngWordCalls, want) {
		t.Errorf("NG word calls = %v, want %v", ngWordCalls, want)
	}
	if want := []string{"new", "old"}; !slices.Equal(deletedWords, want) {
		t.Errorf("deleted words = %v, want %v", deletedWords, want)
	}
}

func TestNGWordUsecase_Moderate_Errors(t *testing.T) {
	boom := errors.New("boom")
	owned := []*model.LivestreamModel{{ID: 10, UserID: 2}}

	tests := []struct {
		name string
		// ownedLivestreams, livestreamErr は配信者自身のライブ配信を検索した結果
		ownedLivestreams []*model.LivestreamModel
		livestreamErr    error
		// deleteErr はライブコメントの削除が返すエラー
		deleteErr error
		// ngWords, ngWordCreateErr は NG ワードの取り直し・登録が返す値
		ngWords         []*model.NGWordModel
		ngWordCreateErr error
		wantErr         error
		wantNGWordCalls []string
	}{
		{
			// 他の配信者のライブ配信・存在しないライブ配信のどちらも、検索結果が空になる
			name:    "not the owner",
			wantErr: ErrNotLivestreamOwner,
		},
		{
			name:          "livestream repository error",
			livestreamErr: boom,
			wantErr:       boom,
		},
		{
			name:             "create fails",
			ownedLivestreams: owned,
			ngWordCreateErr:  boom,
			wantErr:          boom,
			wantNGWordCalls:  []string{"Create"},
		},
		{
			name:             "delete fails",
			ownedLivestreams: owned,
			deleteErr:        boom,
			ngWords:          []*model.NGWordModel{{Word: "bad"}},
			wantErr:          boom,
			wantNGWordCalls:  []string{"Create", "FindAllByLivestreamID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ngWordCalls []string
			var created *model.NGWordModel
			ngWordRepo := newNGWordRepositoryForModerate(t, &ngWordCalls, &created, tt.ngWords, tt.ngWordCreateErr)
			var deletedWords []string
			livecommentRepo := newLivecommentRepositoryForModerate(t, &deletedWords, tt.deleteErr)
			u := NewNGWordUsecase(&fakeTxManager{}, newLivestreamRepositoryForModerate(t, tt.ownedLivestreams, tt.livestreamErr), livecommentRepo, ngWordRepo)
			_, err := u.Moderate(context.Background(), 2, 10, "bad")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !slices.Equal(ngWordCalls, tt.wantNGWordCalls) {
				t.Errorf("NG word calls = %v, want %v", ngWordCalls, tt.wantNGWordCalls)
			}
		})
	}
}
