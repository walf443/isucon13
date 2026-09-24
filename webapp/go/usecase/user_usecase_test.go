package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestUserUsecase_FindByName(t *testing.T) {
	want := &model.User{ID: 1, Name: "alice", Theme: model.ThemeModel{ID: 10, UserID: 1}, IconHash: "abc"}
	userRepo := &fakeUserRepository{userDetails: want}
	u := NewUserUsecase(&fakeTxManager{}, userRepo)

	user, err := u.FindByName(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != want {
		t.Errorf("user = %+v, want %+v", user, want)
	}
	if userRepo.gotName != "alice" {
		t.Errorf("name = %q, want %q", userRepo.gotName, "alice")
	}
}

func TestUserUsecase_FindByID(t *testing.T) {
	want := &model.User{ID: 1, Name: "alice"}
	userRepo := &fakeUserRepository{userDetails: want}
	u := NewUserUsecase(&fakeTxManager{}, userRepo)

	user, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != want {
		t.Errorf("user = %+v, want %+v", user, want)
	}
	if userRepo.gotID != 1 {
		t.Errorf("id = %d, want 1", userRepo.gotID)
	}
}

func TestUserUsecase_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		repoErr error
		check   func(t *testing.T, err error)
	}{
		{
			name:    "user not found",
			repoErr: repository.ErrNotFound,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, ErrUserNotFound) {
					t.Errorf("err = %v, want ErrUserNotFound", err)
				}
			},
		},
		{
			// テーマ欠損などの repository のエラーは 404 にしない
			name:    "unexpected error",
			repoErr: boom,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, boom) || errors.Is(err, ErrUserNotFound) {
					t.Errorf("err = %v, want wrapped %v", err, boom)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(&fakeTxManager{}, &fakeUserRepository{err: tt.repoErr})

			_, err := u.FindByName(context.Background(), "alice")
			tt.check(t, err)
			_, err = u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}
