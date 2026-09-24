package handler

import (
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Theme       Theme  `json:"theme,omitempty"`
	IconHash    string `json:"icon_hash,omitempty"`
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
		IconHash: u.IconHash,
	}
}

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// GET /api/user/:username
func (h *UserHandler) GetUser(c echo.Context) error {
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
