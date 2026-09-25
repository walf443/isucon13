package handler

import (
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type themeResponse struct {
	ID       model.ThemeID `json:"id"`
	DarkMode bool          `json:"dark_mode"`
}

type themeHandler struct {
	themeUsecase usecase.ThemeUsecase
}

func newThemeHandler(themeUsecase usecase.ThemeUsecase) *themeHandler {
	return &themeHandler{themeUsecase: themeUsecase}
}

// 配信者のテーマ取得API
// GET /api/user/:username/theme
func (h *themeHandler) GetStreamerTheme(c echo.Context) error {
	ctx := c.Request().Context()

	username := c.Param("username")

	themeModel, err := h.themeUsecase.FindByUsername(ctx, username)
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, themeResponse{
		ID:       themeModel.ID,
		DarkMode: themeModel.DarkMode,
	})
}
