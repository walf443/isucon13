package handler

import (
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type userStatisticsResponse struct {
	Rank              int64  `json:"rank"`
	ViewersCount      int64  `json:"viewers_count"`
	TotalReactions    int64  `json:"total_reactions"`
	TotalLivecomments int64  `json:"total_livecomments"`
	TotalTip          int64  `json:"total_tip"`
	FavoriteEmoji     string `json:"favorite_emoji"`
}

func newUserStatistics(s *model.UserStatistics) userStatisticsResponse {
	return userStatisticsResponse{
		Rank:              s.Rank,
		ViewersCount:      s.ViewersCount,
		TotalReactions:    s.TotalReactions,
		TotalLivecomments: s.TotalLivecomments,
		TotalTip:          s.TotalTip,
		FavoriteEmoji:     s.FavoriteEmoji,
	}
}

type livestreamStatisticsResponse struct {
	Rank           int64 `json:"rank"`
	ViewersCount   int64 `json:"viewers_count"`
	TotalReactions int64 `json:"total_reactions"`
	TotalReports   int64 `json:"total_reports"`
	MaxTip         int64 `json:"max_tip"`
}

func newLivestreamStatistics(s *model.LivestreamStatistics) livestreamStatisticsResponse {
	return livestreamStatisticsResponse{
		Rank:           s.Rank,
		ViewersCount:   s.ViewersCount,
		TotalReactions: s.TotalReactions,
		TotalReports:   s.TotalReports,
		MaxTip:         s.MaxTip,
	}
}

type statisticsHandler struct {
	statisticsUsecase usecase.StatisticsUsecase
}

func newStatisticsHandler(statisticsUsecase usecase.StatisticsUsecase) *statisticsHandler {
	return &statisticsHandler{statisticsUsecase: statisticsUsecase}
}

// GET /api/user/:username/statistics
func (h *statisticsHandler) GetUserStatistics(c echo.Context) error {
	ctx := c.Request().Context()

	username := c.Param("username")

	stats, err := h.statisticsUsecase.FindUserStatistics(ctx, username)
	if errors.Is(err, usecase.ErrUserNotFound) {
		// 他のエンドポイントと違い 404 ではなく 400 (移行前と同じ)
		return echo.NewHTTPError(http.StatusBadRequest, "not found user that has the given username")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newUserStatistics(stats))
}

// GET /api/livestream/:livestream_id/statistics
func (h *statisticsHandler) GetLivestreamStatistics(c echo.Context) error {
	ctx := c.Request().Context()

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	stats, err := h.statisticsUsecase.FindLivestreamStatistics(ctx, livestreamID)
	if errors.Is(err, usecase.ErrLivestreamNotFound) {
		// 404 ではなく 400 (移行前と同じ)
		return echo.NewHTTPError(http.StatusBadRequest, "cannot get stats of not found livestream")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestreamStatistics(stats))
}
