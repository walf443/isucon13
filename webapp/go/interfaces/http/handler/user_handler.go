package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID          model.UserID `json:"id"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name,omitempty"`
	Description string       `json:"description,omitempty"`
	Theme       Theme        `json:"theme,omitempty"`
	IconHash    string       `json:"icon_hash,omitempty"`
}

func newUser(u *model.User) User {
	return User{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Description: u.Description,
		Theme: Theme{
			ID:       u.Theme.ID,
			DarkMode: u.Theme.DarkMode,
		},
		IconHash: string(u.IconHash),
	}
}

type PostUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// Password is non-hashed password.
	Password string               `json:"password"`
	Theme    PostUserRequestTheme `json:"theme"`
}

type PostUserRequestTheme struct {
	DarkMode bool `json:"dark_mode"`
}

type LoginRequest struct {
	Username string `json:"username"`
	// Password is non-hashed password.
	Password string `json:"password"`
}

type userHandler struct {
	userUsecase usecase.UserUsecase
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func newUserHandler(userUsecase usecase.UserUsecase) *userHandler {
	return &userHandler{userUsecase: userUsecase, now: time.Now}
}

// GET /api/user/:username
func (h *userHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()
	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")

	user, err := h.userUsecase.FindByName(ctx, username)
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newUser(user))
}

// GET /api/user/me
func (h *userHandler) GetMe(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	user, err := h.userUsecase.FindByID(ctx, userID)
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the userid in session")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newUser(user))
}

// ユーザ登録API
// POST /api/register
func (h *userHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := PostUserRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	user, err := h.userUsecase.Register(ctx, usecase.RegisterUserInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Password:    req.Password,
		DarkMode:    req.Theme.DarkMode,
	})
	if reserved, ok := errors.AsType[*usecase.ReservedUsernameError](err); ok {
		return echo.NewHTTPError(http.StatusBadRequest, reserved.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newUser(user))
}

// ユーザログインAPI
// POST /api/login
func (h *userHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := LoginRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	user, err := h.userUsecase.Login(ctx, req.Username, req.Password)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sessionEndAt := h.now().Add(1 * time.Hour)

	sessionID := uuid.NewString()

	sess, err := session.Get(DefaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}

	sess.Options = &sessions.Options{
		Domain: "u.isucon.dev",
		MaxAge: int(60000),
		Path:   "/",
	}
	sess.Values[DefaultSessionIDKey] = sessionID
	// VerifyUserSession などは int64 として取り出すので、model.UserID ではなく int64 で保存する
	sess.Values[DefaultUserIDKey] = int64(user.ID)
	sess.Values[DefaultUsernameKey] = user.Name
	sess.Values[DefaultSessionExpiresKey] = sessionEndAt.Unix()

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}
