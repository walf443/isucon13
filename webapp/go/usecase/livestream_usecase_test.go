package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestLivestreamUsecase_FindByID(t *testing.T) {
	want := &model.Livestream{ID: 1, Title: "stream"}
	repo := &fakeLivestreamRepository{livestream: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo)

	got, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if repo.gotID != 1 {
		t.Errorf("id = %d, want 1", repo.gotID)
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
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{err: tt.repoErr})
			_, err := u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

func TestLivestreamUsecase_FindAllByUserID(t *testing.T) {
	want := []*model.Livestream{{ID: 1}, {ID: 2}}
	repo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo)

	got, err := u.FindAllByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if repo.gotUserID != 42 {
		t.Errorf("userID = %d, want 42", repo.gotUserID)
	}
}

func TestLivestreamUsecase_FindAllByUserID_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{err: boom})

	_, err := u.FindAllByUserID(context.Background(), 42)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivestreamUsecase_FindAllByUsername(t *testing.T) {
	want := []*model.Livestream{{ID: 1}}
	userRepo := &fakeUserRepository{id: 42}
	livestreamRepo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, userRepo, &fakeTagRepository{}, livestreamRepo)

	got, err := u.FindAllByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if userRepo.gotName != "alice" {
		t.Errorf("name = %q, want %q", userRepo.gotName, "alice")
	}
	// ユーザ名から引いた ID で検索する
	if livestreamRepo.gotUserID != 42 {
		t.Errorf("userID = %d, want 42", livestreamRepo.gotUserID)
	}
}

func TestLivestreamUsecase_FindAllByUsername_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name           string
		userRepo       *fakeUserRepository
		livestreamRepo *fakeLivestreamRepository
		wantErr        error
	}{
		{
			name:           "user not found",
			userRepo:       &fakeUserRepository{err: repository.ErrNotFound},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        ErrUserNotFound,
		},
		{
			name:           "user repository error",
			userRepo:       &fakeUserRepository{err: boom},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        boom,
		},
		{
			name:           "livestream repository error",
			userRepo:       &fakeUserRepository{id: 42},
			livestreamRepo: &fakeLivestreamRepository{err: boom},
			wantErr:        boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivestreamUsecase(&fakeTxManager{}, tt.userRepo, &fakeTagRepository{}, tt.livestreamRepo)
			_, err := u.FindAllByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLivestreamUsecase_FindAllByTagName(t *testing.T) {
	want := []*model.Livestream{{ID: 2}, {ID: 1}}
	tagRepo := &fakeTagRepository{ids: []model.TagID{7}}
	livestreamRepo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo)

	got, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if tagRepo.gotName != "ゲーム実況" {
		t.Errorf("tag name = %q", tagRepo.gotName)
	}
	if !slices.Equal(livestreamRepo.gotTagIDs, []model.TagID{7}) {
		t.Errorf("tagIDs = %v, want [7]", livestreamRepo.gotTagIDs)
	}
}

func TestLivestreamUsecase_FindAllByTagName_TagNotFound(t *testing.T) {
	livestreamRepo := &fakeLivestreamRepository{}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{ids: nil}, livestreamRepo)

	got, err := u.FindAllByTagName(context.Background(), "nothing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
	// 空の IN () になるので livestream の検索はしない
	if len(livestreamRepo.calls) != 0 {
		t.Errorf("calls = %v, want none", livestreamRepo.calls)
	}
}

func TestLivestreamUsecase_FindAllByTagName_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name           string
		tagRepo        *fakeTagRepository
		livestreamRepo *fakeLivestreamRepository
	}{
		{name: "tag repository error", tagRepo: &fakeTagRepository{err: boom}, livestreamRepo: &fakeLivestreamRepository{}},
		{name: "livestream repository error", tagRepo: &fakeTagRepository{ids: []model.TagID{7}}, livestreamRepo: &fakeLivestreamRepository{err: boom}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tt.tagRepo, tt.livestreamRepo)
			_, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
		})
	}
}

func TestLivestreamUsecase_FindAll(t *testing.T) {
	limit := int64(5)

	tests := []struct {
		name      string
		limit     *int64
		wantCalls []string
		wantLimit int64
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetails"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*model.Livestream{{ID: 1}}
			livestreamRepo := &fakeLivestreamRepository{livestreams: want}
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo)

			got, err := u.FindAll(context.Background(), tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
			if !slices.Equal(livestreamRepo.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", livestreamRepo.calls, tt.wantCalls)
			}
			if livestreamRepo.gotLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", livestreamRepo.gotLimit, tt.wantLimit)
			}
		})
	}
}

func TestLivestreamUsecase_FindAll_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{err: boom})

	_, err := u.FindAll(context.Background(), nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
