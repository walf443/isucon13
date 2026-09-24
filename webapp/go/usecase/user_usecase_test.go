package usecase

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestUserUsecase_FindByName(t *testing.T) {
	userModel := &model.UserModel{ID: 1, Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "x"}
	theme := &model.ThemeModel{ID: 10, UserID: 1, DarkMode: true}
	fallback := []byte("fallback")

	tests := []struct {
		name         string
		iconRepo     *fakeIconRepository
		wantIconHash string
	}{
		{
			name:         "uses registered icon",
			iconRepo:     &fakeIconRepository{image: []byte("icon")},
			wantIconHash: fmt.Sprintf("%x", sha256.Sum256([]byte("icon"))),
		},
		{
			name:         "uses fallback icon when not registered",
			iconRepo:     &fakeIconRepository{err: repository.ErrNotFound},
			wantIconHash: fmt.Sprintf("%x", sha256.Sum256(fallback)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			themeRepo := &fakeThemeRepository{theme: theme}
			u := NewUserUsecase(&fakeTxManager{}, &fakeUserRepository{user: userModel}, themeRepo, tt.iconRepo, fallback)

			user, err := u.FindByName(context.Background(), "alice")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := &model.User{
				ID:          1,
				Name:        "alice",
				DisplayName: "Alice",
				Description: "hello",
				Theme:       *theme,
				IconHash:    tt.wantIconHash,
			}
			if *user != *want {
				t.Errorf("user = %+v, want %+v", user, want)
			}
			if themeRepo.gotUserID != 1 {
				t.Errorf("theme userID = %d, want 1", themeRepo.gotUserID)
			}
		})
	}
}

func TestUserUsecase_FindByName_Errors(t *testing.T) {
	userModel := &model.UserModel{ID: 1, Name: "alice"}
	theme := &model.ThemeModel{ID: 10, UserID: 1}
	boom := errors.New("boom")

	tests := []struct {
		name      string
		userRepo  *fakeUserRepository
		themeRepo *fakeThemeRepository
		iconRepo  *fakeIconRepository
		check     func(t *testing.T, err error)
	}{
		{
			name:      "user not found",
			userRepo:  &fakeUserRepository{err: repository.ErrNotFound},
			themeRepo: &fakeThemeRepository{},
			iconRepo:  &fakeIconRepository{},
			check: func(t *testing.T, err error) {
				if !errors.Is(err, ErrUserNotFound) {
					t.Errorf("err = %v, want ErrUserNotFound", err)
				}
			},
		},
		{
			// テーマが無いのはデータ不整合なので、ユーザ不在 (404) とは区別する
			name:      "theme not found",
			userRepo:  &fakeUserRepository{user: userModel},
			themeRepo: &fakeThemeRepository{err: repository.ErrNotFound},
			iconRepo:  &fakeIconRepository{},
			check: func(t *testing.T, err error) {
				if err == nil || errors.Is(err, ErrUserNotFound) {
					t.Errorf("err = %v, want non-ErrUserNotFound error", err)
				}
			},
		},
		{
			name:      "icon repository error",
			userRepo:  &fakeUserRepository{user: userModel},
			themeRepo: &fakeThemeRepository{theme: theme},
			iconRepo:  &fakeIconRepository{err: boom},
			check: func(t *testing.T, err error) {
				if !errors.Is(err, boom) {
					t.Errorf("err = %v, want %v", err, boom)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(&fakeTxManager{}, tt.userRepo, tt.themeRepo, tt.iconRepo, []byte("fallback"))
			_, err := u.FindByName(context.Background(), "alice")
			tt.check(t, err)
		})
	}
}
