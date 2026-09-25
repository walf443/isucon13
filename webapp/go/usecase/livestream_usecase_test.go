package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestLivestreamUsecase_FindByID(t *testing.T) {
	want := &model.Livestream{ID: 1, Title: "stream"}
	repo := &fakeLivestreamRepository{
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.LivestreamID) (*model.Livestream, error) {
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return want, nil
		},
	}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo)

	got, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindByID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		repoErr error
		check   func(t *testing.T, err error)
	}{
		{
			name:    "livestream not found",
			repoErr: repository.ErrNotFound,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, ErrLivestreamNotFound) {
					t.Errorf("err = %v, want ErrLivestreamNotFound", err)
				}
			},
		},
		{
			name:    "unexpected error",
			repoErr: boom,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, boom) || errors.Is(err, ErrLivestreamNotFound) {
					t.Errorf("err = %v, want wrapped %v", err, boom)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLivestreamRepository{
				findWithDetailsByID: func(context.Context, repository.Querier, model.LivestreamID) (*model.Livestream, error) {
					return nil, tt.repoErr
				},
			}
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo)
			_, err := u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

// newLivestreamRepositoryFindingAllByUserID は配信者 userID のライブ配信として livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
func newLivestreamRepositoryFindingAllByUserID(t *testing.T, userID model.UserID, livestreams []*model.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllWithDetailsByUserID: func(_ context.Context, _ repository.Querier, gotUserID model.UserID) ([]*model.Livestream, error) {
			if gotUserID != userID {
				t.Errorf("userID = %d, want %d", gotUserID, userID)
			}
			return livestreams, err
		},
	}
}

func TestLivestreamUsecase_FindAllByUserID(t *testing.T) {
	want := []*model.Livestream{{ID: 1}, {ID: 2}}
	repo := newLivestreamRepositoryFindingAllByUserID(t, 42, want, nil)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo)

	got, err := u.FindAllByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByUserID_Error(t *testing.T) {
	boom := errors.New("boom")
	repo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, boom)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo)

	_, err := u.FindAllByUserID(context.Background(), 42)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivestreamUsecase_FindAllByUsername(t *testing.T) {
	want := []*model.Livestream{{ID: 1}}
	// ユーザ名から引いた ID で検索する
	livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, want, nil)
	u := NewLivestreamUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", 42, nil), &fakeTagRepository{}, livestreamRepo)

	got, err := u.FindAllByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByUsername_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// userID, userErr はユーザ名から ID を引いた結果
		userID  model.UserID
		userErr error
		// livestreamsErr はライブ配信の検索が返すエラー
		livestreamsErr error
		wantErr        error
	}{
		{
			name:    "user not found",
			userErr: repository.ErrNotFound,
			wantErr: ErrUserNotFound,
		},
		{
			name:    "user repository error",
			userErr: boom,
			wantErr: boom,
		},
		{
			name:           "livestream repository error",
			userID:         42,
			livestreamsErr: boom,
			wantErr:        boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, tt.livestreamsErr)
			u := NewLivestreamUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", tt.userID, tt.userErr), &fakeTagRepository{}, livestreamRepo)
			_, err := u.FindAllByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// newLivestreamRepositoryFindingAllByTagIDs はタグ 7 のライブ配信として livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
// 呼ばれた回数を calls に数える。
func newLivestreamRepositoryFindingAllByTagIDs(t *testing.T, calls *int, livestreams []*model.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllWithDetailsByTagIDs: func(_ context.Context, _ repository.Querier, tagIDs []model.TagID) ([]*model.Livestream, error) {
			*calls++
			if !slices.Equal(tagIDs, []model.TagID{7}) {
				t.Errorf("tagIDs = %v, want [7]", tagIDs)
			}
			return livestreams, err
		},
	}
}

func TestLivestreamUsecase_FindAllByTagName(t *testing.T) {
	want := []*model.Livestream{{ID: 2}, {ID: 1}}
	tagRepo := &fakeTagRepository{
		findIDsByName: func(_ context.Context, _ repository.Querier, name string) ([]model.TagID, error) {
			if name != "ゲーム実況" {
				t.Errorf("tag name = %q", name)
			}
			return []model.TagID{7}, nil
		},
	}
	var livestreamCalls int
	livestreamRepo := newLivestreamRepositoryFindingAllByTagIDs(t, &livestreamCalls, want, nil)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo)

	got, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByTagName_TagNotFound(t *testing.T) {
	var livestreamCalls int
	livestreamRepo := newLivestreamRepositoryFindingAllByTagIDs(t, &livestreamCalls, nil, nil)
	tagRepo := &fakeTagRepository{
		findIDsByName: func(context.Context, repository.Querier, string) ([]model.TagID, error) { return nil, nil },
	}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo)

	got, err := u.FindAllByTagName(context.Background(), "nothing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
	// 空の IN () になるので livestream の検索はしない
	if livestreamCalls != 0 {
		t.Errorf("livestream calls = %d, want 0", livestreamCalls)
	}
}

func TestLivestreamUsecase_FindAllByTagName_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// tagIDs, tagErr はタグの検索が返す値
		tagIDs []model.TagID
		tagErr error
		// livestreamsErr はライブ配信の検索が返すエラー
		livestreamsErr error
	}{
		{name: "tag repository error", tagErr: boom},
		{name: "livestream repository error", tagIDs: []model.TagID{7}, livestreamsErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagRepo := &fakeTagRepository{
				findIDsByName: func(context.Context, repository.Querier, string) ([]model.TagID, error) { return tt.tagIDs, tt.tagErr },
			}
			var livestreamCalls int
			livestreamRepo := newLivestreamRepositoryFindingAllByTagIDs(t, &livestreamCalls, nil, tt.livestreamsErr)
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo)
			_, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
		})
	}
}

// newLivestreamRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
// limit 付きの場合はその値を gotLimit に取り出す。
func newLivestreamRepositoryForFindAll(calls *[]string, gotLimit *model.Limit, livestreams []*model.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllWithDetails: func(context.Context, repository.Querier) ([]*model.Livestream, error) {
			*calls = append(*calls, "FindAllWithDetails")
			return livestreams, err
		},
		findAllWithDetailsLimited: func(_ context.Context, _ repository.Querier, limit model.Limit) ([]*model.Livestream, error) {
			*calls = append(*calls, "FindAllWithDetailsLimited")
			*gotLimit = limit
			return livestreams, err
		},
	}
}

func TestLivestreamUsecase_FindAll(t *testing.T) {
	limit := model.Limit(5)

	tests := []struct {
		name      string
		limit     *model.Limit
		wantCalls []string
		wantLimit model.Limit
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetails"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*model.Livestream{{ID: 1}}
			var calls []string
			var gotLimit model.Limit
			livestreamRepo := newLivestreamRepositoryForFindAll(&calls, &gotLimit, want, nil)
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo)

			got, err := u.FindAll(context.Background(), tt.limit)
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

func TestLivestreamUsecase_FindAll_Error(t *testing.T) {
	boom := errors.New("boom")
	var calls []string
	var gotLimit model.Limit
	livestreamRepo := newLivestreamRepositoryForFindAll(&calls, &gotLimit, nil, boom)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo)

	_, err := u.FindAll(context.Background(), nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
