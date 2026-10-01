package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"golang.org/x/crypto/bcrypt"
)

// alice と aliceUser は、テーマ・アイコンを埋める前後のユーザ alice。
var (
	alice      = &domain.User{ID: 1, Name: "alice", DisplayName: "Alice", Description: "hi"}
	aliceTheme = &domain.Theme{ID: 10, UserID: 1, DarkMode: true}
	aliceUser  = domain.UserDetail{ID: 1, Name: "alice", DisplayName: "Alice", Description: "hi", Theme: *aliceTheme, IconHash: "default-hash"}
)

func TestUserUsecase_FindByName(t *testing.T) {
	userRepo := &fakeUserRepository{
		findByName: func(_ context.Context, _ repository.Querier, name string) (*domain.User, error) {
			if name != "alice" {
				t.Errorf("name = %q, want %q", name, "alice")
			}
			return alice, nil
		},
	}
	userFiller := newUserFillerForTest(map[domain.UserID]*domain.Theme{1: aliceTheme}, nil, nil)
	u := NewUserUsecase(&fakeTxManager{}, userRepo, userFiller)

	userDetail, err := u.FindByName(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *userDetail != aliceUser {
		t.Errorf("user = %+v, want %+v", *userDetail, aliceUser)
	}
}

func TestUserUsecase_FindByID(t *testing.T) {
	userRepo := &fakeUserRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.UserID) (*domain.User, error) {
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return alice, nil
		},
	}
	userFiller := newUserFillerForTest(map[domain.UserID]*domain.Theme{1: aliceTheme}, nil, nil)
	u := NewUserUsecase(&fakeTxManager{}, userRepo, userFiller)

	userDetail, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *userDetail != aliceUser {
		t.Errorf("user = %+v, want %+v", *userDetail, aliceUser)
	}
}

func TestUserUsecase_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// user, userErr はユーザを引いた結果
		user    *domain.User
		userErr error
		// themes はユーザの ID ごとのテーマ (無ければテーマ欠損)
		themes  map[domain.UserID]*domain.Theme
		wantErr error
		wantMsg string
	}{
		{name: "user not found", userErr: repository.ErrNotFound, wantErr: ErrUserNotFound, wantMsg: ErrUserNotFound.Error()},
		{name: "get user fails", userErr: boom, wantErr: boom, wantMsg: "failed to get user: boom"},
		{
			// テーマ欠損はデータ不整合なので 404 (ErrUserNotFound) にしない
			name:    "theme not found",
			user:    alice,
			wantErr: errMissingDetail,
			wantMsg: "failed to get user: failed to get theme of user 1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &fakeUserRepository{
				findByName: func(context.Context, repository.Querier, string) (*domain.User, error) {
					return tt.user, tt.userErr
				},
				findByID: func(context.Context, repository.Querier, domain.UserID) (*domain.User, error) {
					return tt.user, tt.userErr
				},
			}
			u := NewUserUsecase(&fakeTxManager{}, userRepo, newUserFillerForTest(tt.themes, nil, nil))

			check := func(err error) {
				t.Helper()
				if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
					t.Errorf("err = %v, want %q", err, tt.wantMsg)
				}
				if tt.wantErr != ErrUserNotFound && errors.Is(err, ErrUserNotFound) {
					t.Errorf("err = %v, should not be ErrUserNotFound", err)
				}
			}
			_, err := u.FindByName(context.Background(), "alice")
			check(err)
			_, err = u.FindByID(context.Background(), 1)
			check(err)
		})
	}
}

func TestUserUsecase_Login(t *testing.T) {
	hashed, err := domain.HashPassword("s3cret")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &domain.User{ID: 5, Name: "alice", HashedPassword: hashed}
	boom := errors.New("boom")

	tests := []struct {
		name string
		// user, userErr はユーザ名でユーザを引いた結果
		user     *domain.User
		userErr  error
		password domain.PlainPassword
		want     *domain.User
		wantErr  error
		wantMsg  string
	}{
		{name: "succeeds", user: user, password: "s3cret", want: user},
		{name: "wrong password", user: user, password: "wrong", wantErr: ErrInvalidCredentials},
		{name: "user not found", userErr: repository.ErrNotFound, password: "s3cret", wantErr: ErrInvalidCredentials},
		{name: "get user fails", userErr: boom, password: "s3cret", wantErr: boom, wantMsg: "failed to get user: boom"},
		{
			// ハッシュとして不正な値の場合は 401 ではなくエラーにする (移行前と同じ)
			name:     "broken hash",
			user:     &domain.User{ID: 5, Name: "alice", HashedPassword: "broken"},
			password: "s3cret",
			wantErr:  bcrypt.ErrHashTooShort,
			wantMsg:  "failed to compare hash and password: " + bcrypt.ErrHashTooShort.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := &fakeTxManager{}
			userRepo := &fakeUserRepository{
				findByName: func(_ context.Context, _ repository.Querier, name string) (*domain.User, error) {
					if name != "alice" {
						t.Errorf("name = %q, want %q", name, "alice")
					}
					return tt.user, tt.userErr
				},
			}
			u := NewUserUsecase(txManager, userRepo, nil)
			got, err := u.Login(context.Background(), "alice", tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantMsg != "" && err.Error() != tt.wantMsg {
				t.Errorf("err = %q, want %q", err.Error(), tt.wantMsg)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if txManager.runs != 1 {
				t.Errorf("tx runs = %d, want 1", txManager.runs)
			}
		})
	}
}
