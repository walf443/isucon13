package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_FindByName(t *testing.T) {
	want := &model.User{ID: 1, Name: "alice", Theme: model.ThemeModel{ID: 10, UserID: 1}, IconHash: "abc"}
	userRepo := &fakeUserRepository{userDetails: want}
	u := NewUserUsecase(&fakeTxManager{}, userRepo, &fakeThemeRepository{}, &fakeDNSRecordRegistrar{})

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
	u := NewUserUsecase(&fakeTxManager{}, userRepo, &fakeThemeRepository{}, &fakeDNSRecordRegistrar{})

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
			u := NewUserUsecase(&fakeTxManager{}, &fakeUserRepository{err: tt.repoErr}, &fakeThemeRepository{}, &fakeDNSRecordRegistrar{})

			_, err := u.FindByName(context.Background(), "alice")
			tt.check(t, err)
			_, err = u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

func TestUserUsecase_Register(t *testing.T) {
	want := &model.User{ID: 5, Name: "alice"}
	txManager := &fakeTxManager{}
	userRepo := &fakeUserRepository{createID: 5, userDetails: want}
	var createdTheme *model.ThemeModel
	themeRepo := &fakeThemeRepository{
		create: func(_ context.Context, _ repository.Querier, theme *model.ThemeModel) error {
			createdTheme = theme
			return nil
		},
	}
	dns := &fakeDNSRecordRegistrar{}
	u := NewUserUsecase(txManager, userRepo, themeRepo, dns)

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

	created := userRepo.gotCreated
	if created.Name != "alice" || created.DisplayName != "Alice" || created.Description != "hello" {
		t.Errorf("created = %+v", created)
	}
	// パスワードは bcrypt でハッシュ化して保存する
	if err := bcrypt.CompareHashAndPassword([]byte(created.HashedPassword), []byte("s3cret")); err != nil {
		t.Errorf("hashed password does not match: %v", err)
	}
	if cost, err := bcrypt.Cost([]byte(created.HashedPassword)); err != nil || cost != bcrypt.MinCost {
		t.Errorf("bcrypt cost = %d, %v, want %d", cost, err, bcrypt.MinCost)
	}
	if want := (model.ThemeModel{UserID: 5, DarkMode: true}); createdTheme == nil || *createdTheme != want {
		t.Errorf("theme = %+v, want %+v", createdTheme, want)
	}
	if want := []string{"alice"}; !slices.Equal(dns.gotNames, want) {
		t.Errorf("dns names = %v, want %v", dns.gotNames, want)
	}
	if want := []string{"Create", "FindWithDetailsByID"}; !slices.Equal(userRepo.calls, want) || userRepo.gotID != 5 {
		t.Errorf("user calls = %v, id = %d", userRepo.calls, userRepo.gotID)
	}
	if txManager.runs != 1 {
		t.Errorf("tx runs = %d, want 1", txManager.runs)
	}
}

func TestUserUsecase_Register_ReservedUsername(t *testing.T) {
	txManager := &fakeTxManager{}
	userRepo := &fakeUserRepository{}
	dns := &fakeDNSRecordRegistrar{}
	u := NewUserUsecase(txManager, userRepo, &fakeThemeRepository{}, dns)

	_, err := u.Register(context.Background(), RegisterUserInput{Name: "pipe", Password: "x"})
	if !errors.Is(err, ErrReservedUsername) {
		t.Fatalf("err = %v, want ErrReservedUsername", err)
	}
	// トランザクションを開始する前に弾く (移行前と同じ)
	if txManager.runs != 0 || len(userRepo.calls) != 0 || len(dns.gotNames) != 0 {
		t.Errorf("tx runs = %d, user calls = %v, dns names = %v", txManager.runs, userRepo.calls, dns.gotNames)
	}
}

func TestUserUsecase_Register_Errors(t *testing.T) {
	boom := errors.New("boom")
	// DNS の登録エラーはコマンドの出力とエラーをそのままメッセージにする
	dnsErr := errors.New("Error: zone not found: exit status 1")

	tests := []struct {
		name     string
		userRepo *fakeUserRepository
		// themeCreateErr はテーマの登録が返すエラー
		themeCreateErr error
		dns            *fakeDNSRecordRegistrar
		wantErr        error
		wantMsg        string
		wantCalls      []string
		wantDNS        []string
	}{
		{
			name:      "insert user fails",
			userRepo:  &fakeUserRepository{createErr: boom},
			dns:       &fakeDNSRecordRegistrar{},
			wantErr:   boom,
			wantMsg:   "failed to insert user: boom",
			wantCalls: []string{"Create"},
		},
		{
			name:           "insert theme fails",
			userRepo:       &fakeUserRepository{createID: 5},
			themeCreateErr: boom,
			dns:            &fakeDNSRecordRegistrar{},
			wantErr:        boom,
			wantMsg:        "failed to insert user theme: boom",
			wantCalls:      []string{"Create"},
		},
		{
			name:      "dns registration fails",
			userRepo:  &fakeUserRepository{createID: 5},
			dns:       &fakeDNSRecordRegistrar{err: dnsErr},
			wantErr:   dnsErr,
			wantMsg:   dnsErr.Error(),
			wantCalls: []string{"Create"},
			wantDNS:   []string{"alice"},
		},
		{
			name:      "fill fails",
			userRepo:  &fakeUserRepository{createID: 5, err: boom},
			dns:       &fakeDNSRecordRegistrar{},
			wantErr:   boom,
			wantMsg:   "failed to fill user: boom",
			wantCalls: []string{"Create", "FindWithDetailsByID"},
			wantDNS:   []string{"alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			themeRepo := &fakeThemeRepository{
				create: func(context.Context, repository.Querier, *model.ThemeModel) error { return tt.themeCreateErr },
			}
			u := NewUserUsecase(&fakeTxManager{}, tt.userRepo, themeRepo, tt.dns)
			_, err := u.Register(context.Background(), RegisterUserInput{Name: "alice", Password: "x"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("err = %q, want %q", err.Error(), tt.wantMsg)
			}
			if !slices.Equal(tt.userRepo.calls, tt.wantCalls) {
				t.Errorf("user calls = %v, want %v", tt.userRepo.calls, tt.wantCalls)
			}
			if !slices.Equal(tt.dns.gotNames, tt.wantDNS) {
				t.Errorf("dns names = %v, want %v", tt.dns.gotNames, tt.wantDNS)
			}
		})
	}
}

func TestUserUsecase_Login(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("s3cret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &model.UserModel{ID: 5, Name: "alice", HashedPassword: string(hashed)}
	boom := errors.New("boom")

	tests := []struct {
		name     string
		userRepo *fakeUserRepository
		password string
		want     *model.UserModel
		wantErr  error
		wantMsg  string
	}{
		{name: "succeeds", userRepo: &fakeUserRepository{user: user}, password: "s3cret", want: user},
		{name: "wrong password", userRepo: &fakeUserRepository{user: user}, password: "wrong", wantErr: ErrInvalidCredentials},
		{name: "user not found", userRepo: &fakeUserRepository{err: repository.ErrNotFound}, password: "s3cret", wantErr: ErrInvalidCredentials},
		{name: "get user fails", userRepo: &fakeUserRepository{err: boom}, password: "s3cret", wantErr: boom, wantMsg: "failed to get user: boom"},
		{
			// ハッシュとして不正な値の場合は 401 ではなくエラーにする (移行前と同じ)
			name:     "broken hash",
			userRepo: &fakeUserRepository{user: &model.UserModel{ID: 5, Name: "alice", HashedPassword: "broken"}},
			password: "s3cret",
			wantErr:  bcrypt.ErrHashTooShort,
			wantMsg:  "failed to compare hash and password: " + bcrypt.ErrHashTooShort.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := &fakeTxManager{}
			u := NewUserUsecase(txManager, tt.userRepo, &fakeThemeRepository{}, &fakeDNSRecordRegistrar{})
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
			if tt.userRepo.gotName != "alice" || txManager.runs != 1 {
				t.Errorf("name = %q, tx runs = %d", tt.userRepo.gotName, txManager.runs)
			}
		})
	}
}
