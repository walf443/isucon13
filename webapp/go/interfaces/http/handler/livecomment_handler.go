package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type Livecomment struct {
	ID         model.LivecommentID `json:"id"`
	User       User                `json:"user"`
	Livestream Livestream          `json:"livestream"`
	Comment    string              `json:"comment"`
	Tip        int64               `json:"tip"`
	CreatedAt  int64               `json:"created_at"`
}

func newLivecomment(l *model.Livecomment) Livecomment {
	return Livecomment{
		ID:         l.ID,
		User:       newUser(&l.User),
		Livestream: newLivestream(&l.Livestream),
		Comment:    l.Comment,
		Tip:        l.Tip,
		CreatedAt:  l.CreatedAt,
	}
}

type LivecommentReport struct {
	ID          model.LivecommentReportID `json:"id"`
	Reporter    User                      `json:"reporter"`
	Livecomment Livecomment               `json:"livecomment"`
	CreatedAt   int64                     `json:"created_at"`
}

func newLivecommentReport(r *model.LivecommentReport) LivecommentReport {
	return LivecommentReport{
		ID:          r.ID,
		Reporter:    newUser(&r.Reporter),
		Livecomment: newLivecomment(&r.Livecomment),
		CreatedAt:   r.CreatedAt,
	}
}

type PostLivecommentRequest struct {
	Comment string `json:"comment"`
	Tip     int64  `json:"tip"`
}

type livecommentHandler struct {
	livecommentUsecase usecase.LivecommentUsecase
}

func newLivecommentHandler(livecommentUsecase usecase.LivecommentUsecase) *livecommentHandler {
	return &livecommentHandler{livecommentUsecase: livecommentUsecase}
}

// GET /api/livestream/:livestream_id/livecomment
func (h *livecommentHandler) GetLivecomments(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	limit, err := parseLimitQueryParam(c, maxLivecommentsLimit)
	if err != nil {
		return err
	}

	livecommentModels, err := h.livecommentUsecase.FindAllByLivestreamID(ctx, livestreamID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	livecomments := make([]Livecomment, len(livecommentModels))
	for i, l := range livecommentModels {
		livecomments[i] = newLivecomment(l)
	}
	return c.JSON(http.StatusOK, livecomments)
}

// (配信者向け)ライブコメントの報告一覧取得API
// GET /api/livestream/:livestream_id/report
func (h *livecommentHandler) GetLivecommentReports(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	reportModels, err := h.livecommentUsecase.FindAllReportsByLivestreamID(ctx, userID, livestreamID)
	if errors.Is(err, usecase.ErrNotLivestreamOwner) {
		return echo.NewHTTPError(http.StatusForbidden, "can't get other streamer's livecomment reports")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reports := make([]LivecommentReport, len(reportModels))
	for i, r := range reportModels {
		reports[i] = newLivecommentReport(r)
	}
	return c.JSON(http.StatusOK, reports)
}

// ライブコメント投稿
// POST /api/livestream/:livestream_id/livecomment
func (h *livecommentHandler) PostLivecomment(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	if err := VerifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	var req PostLivecommentRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	livecomment, err := h.livecommentUsecase.Create(ctx, userID, livestreamID, req.Comment, req.Tip)
	if errors.Is(err, usecase.ErrLivestreamNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "livestream not found")
	}
	if errors.Is(err, usecase.ErrSpamLivecomment) {
		return echo.NewHTTPError(http.StatusBadRequest, "このコメントがスパム判定されました")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newLivecomment(livecomment))
}

// ライブコメント報告
// POST /api/livestream/:livestream_id/livecomment/:livecomment_id/report
func (h *livecommentHandler) PostLivecommentReport(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	livecommentID, err := model.ParseLivecommentID(c.Param("livecomment_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livecomment_id in path must be integer")
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	report, err := h.livecommentUsecase.Report(ctx, userID, livestreamID, livecommentID)
	if errors.Is(err, usecase.ErrLivestreamNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "livestream not found")
	}
	if errors.Is(err, usecase.ErrLivecommentNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "livecomment not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newLivecommentReport(report))
}
