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

type userResponse struct {
	ID          model.UserID  `json:"id"`
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name,omitempty"`
	Description string        `json:"description,omitempty"`
	Theme       themeResponse `json:"theme,omitempty"`
	IconHash    string        `json:"icon_hash,omitempty"`
}

func newUser(u *model.User) userResponse {
	return userResponse{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Description: u.Description,
		Theme: themeResponse{
			ID:       u.Theme.ID,
			DarkMode: u.Theme.DarkMode,
		},
		IconHash: string(u.IconHash),
	}
}

type postUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// Password is non-hashed password.
	Password string               `json:"password"`
	Theme    postUserRequestTheme `json:"theme"`
}

type postUserRequestTheme struct {
	DarkMode bool `json:"dark_mode"`
}

type loginRequest struct {
	Username string `json:"username"`
	// Password is non-hashed password.
	Password string `json:"password"`
}

type userHandler struct {
	userUsecase         usecase.UserUsecase
	registrationUsecase usecase.UserRegistrationUsecase
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func newUserHandler(userUsecase usecase.UserUsecase, registrationUsecase usecase.UserRegistrationUsecase) *userHandler {
	return &userHandler{userUsecase: userUsecase, registrationUsecase: registrationUsecase, now: time.Now}
}

// GET /api/user/:username
func (h *userHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()
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

	req := postUserRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	user, err := h.registrationUsecase.Register(ctx, usecase.RegisterUserInput{
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

	req := loginRequest{}
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

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}

	sess.Options = &sessions.Options{
		Domain: "u.isucon.dev",
		MaxAge: int(60000),
		Path:   "/",
	}
	sess.Values[defaultSessionIDKey] = sessionID
	// requireSession などは int64 として取り出すので、model.UserID ではなく int64 で保存する
	sess.Values[defaultUserIDKey] = int64(user.ID)
	sess.Values[defaultUsernameKey] = user.Name
	sess.Values[defaultSessionExpiresKey] = sessionEndAt.Unix()

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}
