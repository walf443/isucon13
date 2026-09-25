package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_FindByName(t *testing.T) {
	want := &model.User{ID: 1, Name: "alice", Theme: model.ThemeModel{ID: 10, UserID: 1}, IconHash: "abc"}
	userRepo := &fakeUserRepository{
		findWithDetailsByName: func(_ context.Context, _ repository.Querier, name string) (*model.User, error) {
			if name != "alice" {
				t.Errorf("name = %q, want %q", name, "alice")
			}
			return want, nil
		},
	}
	u := NewUserUsecase(&fakeTxManager{}, userRepo)

	user, err := u.FindByName(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != want {
		t.Errorf("user = %+v, want %+v", user, want)
	}
}

func TestUserUsecase_FindByID(t *testing.T) {
	want := &model.User{ID: 1, Name: "alice"}
	userRepo := &fakeUserRepository{
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.UserID) (*model.User, error) {
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return want, nil
		},
	}
	u := NewUserUsecase(&fakeTxManager{}, userRepo)

	user, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != want {
		t.Errorf("user = %+v, want %+v", user, want)
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
			userRepo := &fakeUserRepository{
				findWithDetailsByName: func(context.Context, repository.Querier, string) (*model.User, error) { return nil, tt.repoErr },
				findWithDetailsByID:   func(context.Context, repository.Querier, model.UserID) (*model.User, error) { return nil, tt.repoErr },
			}
			u := NewUserUsecase(&fakeTxManager{}, userRepo)

			_, err := u.FindByName(context.Background(), "alice")
			tt.check(t, err)
			_, err = u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

func TestUserUsecase_Login(t *testing.T) {
	hashed, err := model.HashPassword("s3cret")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &model.UserModel{ID: 5, Name: "alice", HashedPassword: hashed}
	boom := errors.New("boom")

	tests := []struct {
		name string
		// user, userErr はユーザ名でユーザを引いた結果
		user     *model.UserModel
		userErr  error
		password string
		want     *model.UserModel
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
			user:     &model.UserModel{ID: 5, Name: "alice", HashedPassword: "broken"},
			password: "s3cret",
			wantErr:  bcrypt.ErrHashTooShort,
			wantMsg:  "failed to compare hash and password: " + bcrypt.ErrHashTooShort.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := &fakeTxManager{}
			userRepo := &fakeUserRepository{
				findByName: func(_ context.Context, _ repository.Querier, name string) (*model.UserModel, error) {
					if name != "alice" {
						t.Errorf("name = %q, want %q", name, "alice")
					}
					return tt.user, tt.userErr
				},
			}
			u := NewUserUsecase(txManager, userRepo)
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
