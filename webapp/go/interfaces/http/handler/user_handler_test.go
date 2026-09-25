package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type fakeUserUsecase struct {
	user      *model.User
	userModel *model.UserModel
	err       error
	gotID     model.UserID
	gotName   string

	gotInput    usecase.RegisterUserInput
	gotPassword string
}

func (u *fakeUserUsecase) Register(ctx context.Context, input usecase.RegisterUserInput) (*model.User, error) {
	u.gotInput = input
	return u.user, u.err
}

func (u *fakeUserUsecase) Login(ctx context.Context, username string, password string) (*model.UserModel, error) {
	u.gotName = username
	u.gotPassword = password
	return u.userModel, u.err
}

func (u *fakeUserUsecase) FindByID(ctx context.Context, id model.UserID) (*model.User, error) {
	u.gotID = id
	return u.user, u.err
}

func (u *fakeUserUsecase) FindByName(ctx context.Context, name string) (*model.User, error) {
	u.gotName = name
	return u.user, u.err
}

func TestUserHandler_GetUser(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns user",
			cookie: validCookie,
			usecase: &fakeUserUsecase{user: &model.User{
				ID:          1,
				Name:        "alice",
				DisplayName: "Alice",
				Description: "hello",
				Theme:       model.ThemeModel{ID: 10, UserID: 1, DarkMode: true},
				IconHash:    "abc",
			}},
			wantCode: http.StatusOK,
			wantBody: `{"id":1,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":10,"dark_mode":true},"icon_hash":"abc"}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeUserUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 404 when user is not found",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/:username", newUserHandler(tt.usecase).GetUser)

			req := httptest.NewRequest(http.MethodGet, "/api/user/alice", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie(t))
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusOK && tt.usecase.gotName != "alice" {
				t.Errorf("name = %q, want %q", tt.usecase.gotName, "alice")
			}
		})
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 42, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns logged-in user",
			cookie: validCookie,
			usecase: &fakeUserUsecase{user: &model.User{
				ID:          42,
				Name:        "alice",
				DisplayName: "Alice",
				Description: "hello",
				Theme:       model.ThemeModel{ID: 10, UserID: 42, DarkMode: false},
				IconHash:    "abc",
			}},
			wantCode: http.StatusOK,
			wantBody: `{"id":42,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":10,"dark_mode":false},"icon_hash":"abc"}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeUserUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 404 when session user is not found",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/me", newUserHandler(tt.usecase).GetMe)

			req := httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie(t))
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusOK && tt.usecase.gotID != 42 {
				t.Errorf("id = %d, want 42", tt.usecase.gotID)
			}
		})
	}
}

func TestUserHandler_Register(t *testing.T) {
	user := &model.User{ID: 5, Name: "alice", DisplayName: "Alice", Description: "hello", Theme: model.ThemeModel{ID: 9, UserID: 5, DarkMode: true}, IconHash: "abc"}
	reqBody := `{"name":"alice","display_name":"Alice","description":"hello","password":"s3cret","theme":{"dark_mode":true}}`

	tests := []struct {
		name     string
		body     string
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "registers user",
			body:     reqBody,
			usecase:  &fakeUserUsecase{user: user},
			wantCode: http.StatusCreated,
			wantBody: `{"id":5,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":9,"dark_mode":true},"icon_hash":"abc"}` + "\n",
		},
		{
			name:     "returns 400 on invalid json",
			body:     `{`,
			usecase:  &fakeUserUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"failed to decode the request body as json"}` + "\n",
		},
		{
			name:     "returns 400 for reserved username",
			body:     `{"name":"pipe","password":"x"}`,
			usecase:  &fakeUserUsecase{err: usecase.ErrReservedUsername},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"the username 'pipe' is reserved"}` + "\n",
		},
		{
			// DNS の登録エラーなどは usecase のメッセージをそのまま返す
			name:     "returns 500 on unexpected error",
			body:     reqBody,
			usecase:  &fakeUserUsecase{err: errors.New("Error: zone not found: exit status 1")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"Error: zone not found: exit status 1"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.POST("/api/register", newUserHandler(tt.usecase).Register)

			// セッションは不要
			req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusCreated {
				want := usecase.RegisterUserInput{Name: "alice", DisplayName: "Alice", Description: "hello", Password: "s3cret", DarkMode: true}
				if tt.usecase.gotInput != want {
					t.Errorf("input = %+v, want %+v", tt.usecase.gotInput, want)
				}
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	user := &model.UserModel{ID: 5, Name: "alice"}
	// 発行したセッションで認証が通ることも確認するので、実時刻に固定する
	now := time.Unix(time.Now().Unix(), 0)

	tests := []struct {
		name     string
		body     string
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "logs in",
			body:     `{"username":"alice","password":"s3cret"}`,
			usecase:  &fakeUserUsecase{userModel: user},
			wantCode: http.StatusOK,
		},
		{
			name:     "returns 400 on invalid json",
			body:     `{`,
			usecase:  &fakeUserUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"failed to decode the request body as json"}` + "\n",
		},
		{
			name:     "returns 401 on invalid credentials",
			body:     `{"username":"alice","password":"wrong"}`,
			usecase:  &fakeUserUsecase{err: usecase.ErrInvalidCredentials},
			wantCode: http.StatusUnauthorized,
			wantBody: `{"message":"invalid username or password"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			body:     `{"username":"alice","password":"s3cret"}`,
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"boom"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			h := newUserHandler(tt.usecase)
			h.now = func() time.Time { return now }
			e.POST("/api/login", h.Login)

			req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
			cookies := rec.Result().Cookies()
			if tt.wantCode != http.StatusOK {
				if len(cookies) != 0 {
					t.Errorf("cookies = %v, want none", cookies)
				}
				return
			}

			if tt.usecase.gotName != "alice" || tt.usecase.gotPassword != "s3cret" {
				t.Errorf("username = %q, password = %q", tt.usecase.gotName, tt.usecase.gotPassword)
			}
			if len(cookies) != 1 {
				t.Fatalf("len(cookies) = %d, want 1", len(cookies))
			}
			cookie := cookies[0]
			if cookie.Name != DefaultSessionIDKey || cookie.Domain != "u.isucon.dev" || cookie.Path != "/" || cookie.MaxAge != 60000 {
				t.Errorf("cookie = %+v", cookie)
			}

			// 発行したセッションを読み直して中身を確認する
			sessReq := httptest.NewRequest(http.MethodGet, "/", nil)
			sessReq.AddCookie(cookie)
			sess, err := testSessionStore.Get(sessReq, DefaultSessionIDKey)
			if err != nil {
				t.Fatalf("failed to decode session: %v", err)
			}
			// VerifyUserSession などが int64 として取り出せること (model.UserID のままだと認証が全て失敗する)
			if userID, ok := sess.Values[DefaultUserIDKey].(int64); !ok || userID != 5 {
				t.Errorf("USERID = %#v, want int64(5)", sess.Values[DefaultUserIDKey])
			}
			if sess.Values[DefaultUsernameKey] != "alice" {
				t.Errorf("USERNAME = %#v, want alice", sess.Values[DefaultUsernameKey])
			}
			if sess.Values[DefaultSessionExpiresKey] != now.Add(time.Hour).Unix() {
				t.Errorf("EXPIRES = %#v, want %d", sess.Values[DefaultSessionExpiresKey], now.Add(time.Hour).Unix())
			}
			if sid, ok := sess.Values[DefaultSessionIDKey].(string); !ok || sid == "" {
				t.Errorf("SESSIONID = %#v, want non-empty string", sess.Values[DefaultSessionIDKey])
			}

			// 発行したセッションで認証が通ること
			verifyEcho := newTestEcho()
			verifyEcho.GET("/verify", func(c echo.Context) error {
				if err := VerifyUserSession(c); err != nil {
					return err
				}
				userID, err := getSessionUserID(c)
				if err != nil {
					return err
				}
				return c.JSON(http.StatusOK, userID)
			})
			verifyReq := httptest.NewRequest(http.MethodGet, "/verify", nil)
			verifyReq.AddCookie(cookie)
			verifyRec := httptest.NewRecorder()
			verifyEcho.ServeHTTP(verifyRec, verifyReq)
			if verifyRec.Code != http.StatusOK || verifyRec.Body.String() != "5\n" {
				t.Errorf("verify: code = %d, body = %s", verifyRec.Code, verifyRec.Body.String())
			}
		})
	}
}
