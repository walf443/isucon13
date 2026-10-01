package handler

import (
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type livestreamViewerHandler struct {
	viewerUsecase usecase.LivestreamViewerUsecase
}

func newLivestreamViewerHandler(viewerUsecase usecase.LivestreamViewerUsecase) *livestreamViewerHandler {
	return &livestreamViewerHandler{viewerUsecase: viewerUsecase}
}

// ユーザ視聴開始 (viewer)
// POST /api/livestream/:livestream_id/enter
func (h *livestreamViewerHandler) EnterLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	livestreamID, err := domain.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		// 他のエンドポイントと違い "in path" が無い (移行前と同じ)
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id must be integer")
	}

	if err := h.viewerUsecase.Enter(ctx, userID, livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// ユーザ視聴終了 (viewer)
// DELETE /api/livestream/:livestream_id/exit
func (h *livestreamViewerHandler) ExitLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	livestreamID, err := domain.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if err := h.viewerUsecase.Exit(ctx, userID, livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
