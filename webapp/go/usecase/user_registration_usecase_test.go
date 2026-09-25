package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newUserRepositoryForRegister は Register で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeUserRepository を返す。
// 登録したユーザは created に取り出し、ID 5 で読み直したユーザとして user を返す。
func newUserRepositoryForRegister(t *testing.T, calls *[]string, created **model.UserModel, user *model.User, createErr, fillErr error) *fakeUserRepository {
	return &fakeUserRepository{
		create: func(_ context.Context, _ repository.Querier, u *model.UserModel) (model.UserID, error) {
			*calls = append(*calls, "Create")
			*created = u
			return 5, createErr
		},
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.UserID) (*model.User, error) {
			*calls = append(*calls, "FindWithDetailsByID")
			if id != 5 {
				t.Errorf("re-read id = %d, want 5", id)
			}
			return user, fillErr
		},
	}
}

func TestUserRegistrationUsecase_Register(t *testing.T) {
	want := &model.User{ID: 5, Name: "alice"}
	txManager := &fakeTxManager{}
	var userCalls []string
	var created *model.UserModel
	userRepo := newUserRepositoryForRegister(t, &userCalls, &created, want, nil, nil)
	var createdTheme *model.ThemeModel
	themeRepo := &fakeThemeRepository{
		create: func(_ context.Context, _ repository.Querier, theme *model.ThemeModel) error {
			createdTheme = theme
			return nil
		},
	}
	var dnsNames []string
	dns := &fakeDNSRecordRegistrar{
		addRecord: func(name string) error {
			dnsNames = append(dnsNames, name)
			return nil
		},
	}
	u := NewUserRegistrationUsecase(txManager, userRepo, themeRepo, dns)

	got, err := u.Register(context.Background(), RegisterUserInput{
		Name:        "alice",
		DisplayName: "Alice",
		Description: "hello",
		Password:    "s3cret",
		DarkMode:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if created == nil {
		t.Fatal("user was not created")
	}
	if created.Name != "alice" || created.DisplayName != "Alice" || created.Description != "hello" {
		t.Errorf("created = %+v", created)
	}
	// パスワードはハッシュ化して保存する (ハッシュ化の方式は model.HashPassword のテストで確認している)
	if matched, err := created.HashedPassword.Matches("s3cret"); err != nil || !matched {
		t.Errorf("hashed password does not match: %v, %v", matched, err)
	}
	if want := (model.ThemeModel{UserID: 5, DarkMode: true}); createdTheme == nil || *createdTheme != want {
		t.Errorf("theme = %+v, want %+v", createdTheme, want)
	}
	if want := []string{"alice"}; !slices.Equal(dnsNames, want) {
		t.Errorf("dns names = %v, want %v", dnsNames, want)
	}
	if want := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(userCalls, want) {
		t.Errorf("user calls = %v, want %v", userCalls, want)
	}
	if txManager.runs != 1 {
		t.Errorf("tx runs = %d, want 1", txManager.runs)
	}
}

func TestUserRegistrationUsecase_Register_ReservedUsername(t *testing.T) {
	txManager := &fakeTxManager{}
	var userCalls []string
	var created *model.UserModel
	userRepo := newUserRepositoryForRegister(t, &userCalls, &created, nil, nil, nil)
	dns := &fakeDNSRecordRegistrar{
		addRecord: func(name string) error {
			t.Errorf("AddRecord(%q) should not be called", name)
			return nil
		},
	}
	u := NewUserRegistrationUsecase(txManager, userRepo, &fakeThemeRepository{}, dns)

	_, err := u.Register(context.Background(), RegisterUserInput{Name: "pipe", Password: "x"})
	reserved, ok := errors.AsType[*ReservedUsernameError](err)
	if !ok {
		t.Fatalf("err = %v, want *ReservedUsernameError", err)
	}
	if reserved.Name != "pipe" {
		t.Errorf("reserved name = %q, want %q", reserved.Name, "pipe")
	}
	// メッセージは移行前の 400 のメッセージと同じ
	if want := "the username 'pipe' is reserved"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
	// トランザクションを開始する前に弾く (移行前と同じ)
	if txManager.runs != 0 || len(userCalls) != 0 {
		t.Errorf("tx runs = %d, user calls = %v", txManager.runs, userCalls)
	}
}

func TestUserRegistrationUsecase_Register_Errors(t *testing.T) {
	boom := errors.New("boom")
	// DNS の登録エラーはコマンドの出力とエラーをそのままメッセージにする
	dnsErr := errors.New("Error: zone not found: exit status 1")

	tests := []struct {
		name string
		// userCreateErr, userFillErr はユーザの登録・取り直しが返すエラー
		userCreateErr error
		userFillErr   error
		// themeCreateErr はテーマの登録が返すエラー
		themeCreateErr error
		// dnsErr は DNS の登録が返すエラー
		dnsErr    error
		wantErr   error
		wantMsg   string
		wantCalls []string
		wantDNS   []string
	}{
		{
			name:          "insert user fails",
			userCreateErr: boom,
			wantErr:       boom,
			wantMsg:       "failed to insert user: boom",
			wantCalls:     []string{"Create"},
		},
		{
			name:           "insert theme fails",
			themeCreateErr: boom,
			wantErr:        boom,
			wantMsg:        "failed to insert user theme: boom",
			wantCalls:      []string{"Create"},
		},
		{
			name:      "dns registration fails",
			dnsErr:    dnsErr,
			wantErr:   dnsErr,
			wantMsg:   dnsErr.Error(),
			wantCalls: []string{"Create"},
			wantDNS:   []string{"alice"},
		},
		{
			name:        "fill fails",
			userFillErr: boom,
			wantErr:     boom,
			wantMsg:     "failed to fill user: boom",
			wantCalls:   []string{"Create", "FindWithDetailsByID"},
			wantDNS:     []string{"alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			themeRepo := &fakeThemeRepository{
				create: func(context.Context, repository.Querier, *model.ThemeModel) error { return tt.themeCreateErr },
			}
			var dnsNames []string
			dns := &fakeDNSRecordRegistrar{
				addRecord: func(name string) error {
					dnsNames = append(dnsNames, name)
					return tt.dnsErr
				},
			}
			var userCalls []string
			var created *model.UserModel
			userRepo := newUserRepositoryForRegister(t, &userCalls, &created, nil, tt.userCreateErr, tt.userFillErr)
			u := NewUserRegistrationUsecase(&fakeTxManager{}, userRepo, themeRepo, dns)
			_, err := u.Register(context.Background(), RegisterUserInput{Name: "alice", Password: "x"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("err = %q, want %q", err.Error(), tt.wantMsg)
			}
			if !slices.Equal(userCalls, tt.wantCalls) {
				t.Errorf("user calls = %v, want %v", userCalls, tt.wantCalls)
			}
			if !slices.Equal(dnsNames, tt.wantDNS) {
				t.Errorf("dns names = %v, want %v", dnsNames, tt.wantDNS)
			}
		})
	}
}
