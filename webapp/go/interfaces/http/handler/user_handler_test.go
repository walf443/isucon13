package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

	gotPassword string
}

type fakeUserRegistrationUsecase struct {
	user *model.User
	err  error

	gotInput usecase.RegisterUserInput
}

func (u *fakeUserRegistrationUsecase) Register(ctx context.Context, input usecase.RegisterUserInput) (*model.User, error) {
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
	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns user",
			cookie: sessionAs(1),
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
			name:     "returns 404 when user is not found",
			cookie:   sessionAs(1),
			usecase:  &fakeUserUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
			wantBody: errorBody(http.StatusNotFound, "not found user that has the given username"),
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   sessionAs(1),
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newUserHandler(tt.usecase, nil).GetUser, testRequest{method: http.MethodGet, route: "/api/user/:username", path: "/api/user/alice", cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusOK && tt.usecase.gotName != "alice" {
				t.Errorf("name = %q, want %q", tt.usecase.gotName, "alice")
			}
		})
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns logged-in user",
			cookie: sessionAs(42),
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
			name:     "returns 404 when session user is not found",
			cookie:   sessionAs(42),
			usecase:  &fakeUserUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
			wantBody: errorBody(http.StatusNotFound, "not found user that has the userid in session"),
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   sessionAs(42),
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newUserHandler(tt.usecase, nil).GetMe, testRequest{method: http.MethodGet, route: "/api/user/me", path: "/api/user/me", cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
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
		usecase  *fakeUserRegistrationUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "registers user",
			body:     reqBody,
			usecase:  &fakeUserRegistrationUsecase{user: user},
			wantCode: http.StatusCreated,
			wantBody: `{"id":5,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":9,"dark_mode":true},"icon_hash":"abc"}` + "\n",
		},
		{
			name:     "returns 400 on invalid json",
			body:     `{`,
			usecase:  &fakeUserRegistrationUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "failed to decode the request body as json"),
		},
		{
			name:     "returns 400 for reserved username",
			body:     `{"name":"pipe","password":"x"}`,
			usecase:  &fakeUserRegistrationUsecase{err: &usecase.ReservedUsernameError{Name: "pipe"}},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "the username 'pipe' is reserved"),
		},
		{
			// メッセージには usecase のエラーにある予約済みの名前を使い、リクエストのユーザ名はそのまま返さない
			name:     "uses the reserved name from the error, not the request",
			body:     `{"name":"<b>ADMIN</b>","password":"x"}`,
			usecase:  &fakeUserRegistrationUsecase{err: &usecase.ReservedUsernameError{Name: "admin"}},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "the username 'admin' is reserved"),
		},
		{
			// DNS の登録エラーなどは usecase のメッセージをそのまま返す
			name:     "returns 500 on unexpected error",
			body:     reqBody,
			usecase:  &fakeUserRegistrationUsecase{err: errors.New("Error: zone not found: exit status 1")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "Error: zone not found: exit status 1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セッションは不要
			rec := serve(t, newUserHandler(nil, tt.usecase).Register, testRequest{method: http.MethodPost, route: "/api/register", path: "/api/register", body: tt.body})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
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
			wantBody: errorBody(http.StatusBadRequest, "failed to decode the request body as json"),
		},
		{
			name:     "returns 401 on invalid credentials",
			body:     `{"username":"alice","password":"wrong"}`,
			usecase:  &fakeUserUsecase{err: usecase.ErrInvalidCredentials},
			wantCode: http.StatusUnauthorized,
			wantBody: errorBody(http.StatusUnauthorized, "invalid username or password"),
		},
		{
			name:     "returns 500 on unexpected error",
			body:     `{"username":"alice","password":"s3cret"}`,
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUserHandler(tt.usecase, nil)
			h.now = func() time.Time { return now }
			rec := serve(t, h.Login, testRequest{method: http.MethodPost, route: "/api/login", path: "/api/login", body: tt.body})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
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
			if cookie.Name != defaultSessionIDKey || cookie.Domain != "u.isucon.dev" || cookie.Path != "/" || cookie.MaxAge != 60000 {
				t.Errorf("cookie = %+v", cookie)
			}

			// 発行したセッションを読み直して中身を確認する
			sessReq := httptest.NewRequest(http.MethodGet, "/", nil)
			sessReq.AddCookie(cookie)
			sess, err := testSessionStore.Get(sessReq, defaultSessionIDKey)
			if err != nil {
				t.Fatalf("failed to decode session: %v", err)
			}
			// requireSession などが int64 として取り出せること (model.UserID のままだと認証が全て失敗する)
			if userID, ok := sess.Values[defaultUserIDKey].(int64); !ok || userID != 5 {
				t.Errorf("USERID = %#v, want int64(5)", sess.Values[defaultUserIDKey])
			}
			if sess.Values[defaultUsernameKey] != "alice" {
				t.Errorf("USERNAME = %#v, want alice", sess.Values[defaultUsernameKey])
			}
			if sess.Values[defaultSessionExpiresKey] != now.Add(time.Hour).Unix() {
				t.Errorf("EXPIRES = %#v, want %d", sess.Values[defaultSessionExpiresKey], now.Add(time.Hour).Unix())
			}
			if sid, ok := sess.Values[defaultSessionIDKey].(string); !ok || sid == "" {
				t.Errorf("SESSIONID = %#v, want non-empty string", sess.Values[defaultSessionIDKey])
			}

			// 発行したセッションで認証が通ること
			verifyEcho := newTestEcho()
			verifyEcho.GET("/verify", func(c echo.Context) error {
				userID, err := getSessionUserID(c)
				if err != nil {
					return err
				}
				return c.JSON(http.StatusOK, userID)
			}, requireSession(func() time.Time { return now }))
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
