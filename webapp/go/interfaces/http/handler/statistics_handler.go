package handler

import (
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type UserStatistics struct {
	Rank              int64  `json:"rank"`
	ViewersCount      int64  `json:"viewers_count"`
	TotalReactions    int64  `json:"total_reactions"`
	TotalLivecomments int64  `json:"total_livecomments"`
	TotalTip          int64  `json:"total_tip"`
	FavoriteEmoji     string `json:"favorite_emoji"`
}

func newUserStatistics(s *model.UserStatistics) UserStatistics {
	return UserStatistics{
		Rank:              s.Rank,
		ViewersCount:      s.ViewersCount,
		TotalReactions:    s.TotalReactions,
		TotalLivecomments: s.TotalLivecomments,
		TotalTip:          s.TotalTip,
		FavoriteEmoji:     s.FavoriteEmoji,
	}
}

type StatisticsHandler struct {
	statisticsUsecase usecase.StatisticsUsecase
}

func NewStatisticsHandler(statisticsUsecase usecase.StatisticsUsecase) *StatisticsHandler {
	return &StatisticsHandler{statisticsUsecase: statisticsUsecase}
}

// GET /api/user/:username/statistics
func (h *StatisticsHandler) GetUserStatistics(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

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
