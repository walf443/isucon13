package handler

import (
	"net/http"
	"strconv"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type LivestreamViewerHandler struct {
	viewerUsecase usecase.LivestreamViewerUsecase
}

func NewLivestreamViewerHandler(viewerUsecase usecase.LivestreamViewerUsecase) *LivestreamViewerHandler {
	return &LivestreamViewerHandler{viewerUsecase: viewerUsecase}
}

// ユーザ視聴開始 (viewer)
// POST /api/livestream/:livestream_id/enter
func (h *LivestreamViewerHandler) EnterLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		// 他のエンドポイントと違い "in path" が無い (移行前と同じ)
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id must be integer")
	}

	if err := h.viewerUsecase.Enter(ctx, userID, model.LivestreamID(livestreamID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// ユーザ視聴終了 (viewer)
// DELETE /api/livestream/:livestream_id/exit
func (h *LivestreamViewerHandler) ExitLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if err := h.viewerUsecase.Exit(ctx, userID, model.LivestreamID(livestreamID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
