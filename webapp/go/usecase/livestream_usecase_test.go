package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestLivestreamUsecase_FindByID(t *testing.T) {
	want := &model.Livestream{ID: 1, Title: "stream"}
	repo := &fakeLivestreamRepository{livestream: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, repo)

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
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeLivestreamRepository{err: tt.repoErr})
			_, err := u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

func TestLivestreamUsecase_FindAllByUserID(t *testing.T) {
	want := []*model.Livestream{{ID: 1}, {ID: 2}}
	repo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, repo)

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
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeLivestreamRepository{err: boom})

	_, err := u.FindAllByUserID(context.Background(), 42)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivestreamUsecase_FindAllByUsername(t *testing.T) {
	want := []*model.Livestream{{ID: 1}}
	userRepo := &fakeUserRepository{id: 42}
	livestreamRepo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, userRepo, livestreamRepo)

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
			u := NewLivestreamUsecase(&fakeTxManager{}, tt.userRepo, tt.livestreamRepo)
			_, err := u.FindAllByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
