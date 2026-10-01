package handler

import (
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type themeResponse struct {
	ID       domain.ThemeID `json:"id"`
	DarkMode bool           `json:"dark_mode"`
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

	username, err := domain.ParseUsername(c.Param("username"))
	if err != nil {
		// 不正な形のユーザ名のユーザは存在しないので、存在しない場合と同じ応答にする
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}

	theme, err := h.themeUsecase.FindByUsername(ctx, username)
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, themeResponse{
		ID:       theme.ID,
		DarkMode: theme.DarkMode,
	})
}
