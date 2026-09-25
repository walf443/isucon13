package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type livestreamResponse struct {
	ID           domain.LivestreamID `json:"id"`
	Owner        userResponse        `json:"owner"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	PlaylistUrl  string              `json:"playlist_url"`
	ThumbnailUrl string              `json:"thumbnail_url"`
	Tags         []tagResponse       `json:"tags"`
	StartAt      int64               `json:"start_at"`
	EndAt        int64               `json:"end_at"`
}

func newLivestream(l *domain.Livestream) livestreamResponse {
	tags := make([]tagResponse, len(l.Tags))
	for i, tag := range l.Tags {
		tags[i] = tagResponse{ID: tag.ID, Name: tag.Name}
	}
	return livestreamResponse{
		ID:           l.ID,
		Owner:        newUser(&l.Owner),
		Title:        l.Title,
		Description:  l.Description,
		PlaylistUrl:  l.PlaylistUrl,
		ThumbnailUrl: l.ThumbnailUrl,
		Tags:         tags,
		StartAt:      l.StartAt,
		EndAt:        l.EndAt,
	}
}

func newLivestreams(ls []*domain.Livestream) []livestreamResponse {
	livestreams := make([]livestreamResponse, len(ls))
	for i, l := range ls {
		livestreams[i] = newLivestream(l)
	}
	return livestreams
}

type reserveLivestreamRequest struct {
	Tags         []int64 `json:"tags"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	PlaylistUrl  string  `json:"playlist_url"`
	ThumbnailUrl string  `json:"thumbnail_url"`
	StartAt      int64   `json:"start_at"`
	EndAt        int64   `json:"end_at"`
}

type livestreamHandler struct {
	livestreamUsecase  usecase.LivestreamUsecase
	reservationUsecase usecase.LivestreamReservationUsecase
}

func newLivestreamHandler(livestreamUsecase usecase.LivestreamUsecase, reservationUsecase usecase.LivestreamReservationUsecase) *livestreamHandler {
	return &livestreamHandler{livestreamUsecase: livestreamUsecase, reservationUsecase: reservationUsecase}
}

// GET /api/livestream/:livestream_id
func (h *livestreamHandler) GetLivestream(c echo.Context) error {
	ctx := c.Request().Context()

	livestreamID, err := domain.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	livestream, err := h.livestreamUsecase.FindByID(ctx, livestreamID)
	if errors.Is(err, usecase.ErrLivestreamNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found livestream that has the given id")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestream(livestream))
}

// GET /api/livestream
func (h *livestreamHandler) GetMyLivestreams(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	livestreams, err := h.livestreamUsecase.FindAllByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestreams(livestreams))
}

// GET /api/user/:username/livestream
func (h *livestreamHandler) GetUserLivestreams(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Param("username")

	livestreams, err := h.livestreamUsecase.FindAllByUsername(ctx, username)
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestreams(livestreams))
}

// GET /api/livestream/search
func (h *livestreamHandler) SearchLivestreams(c echo.Context) error {
	ctx := c.Request().Context()

	var livestreams []*domain.Livestream
	var err error
	if tagName := c.QueryParam("tag"); tagName != "" {
		// タグによる取得 (limit は無視する)
		livestreams, err = h.livestreamUsecase.FindAllByTagName(ctx, tagName)
	} else {
		// 検索条件なし
		var limit *domain.Limit
		limit, err = parseLimitQueryParam(c, maxLivestreamsLimit)
		if err != nil {
			return err
		}
		livestreams, err = h.livestreamUsecase.FindAll(ctx, limit)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestreams(livestreams))
}

// POST /api/livestream/reservation
func (h *livestreamHandler) ReserveLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	var req reserveLivestreamRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	tagIDs := make([]domain.TagID, len(req.Tags))
	for i, tagID := range req.Tags {
		tagIDs[i] = domain.TagID(tagID)
	}

	livestream, err := h.reservationUsecase.Reserve(ctx, userID, usecase.ReserveLivestreamInput{
		TagIDs:       tagIDs,
		Title:        req.Title,
		Description:  req.Description,
		PlaylistUrl:  req.PlaylistUrl,
		ThumbnailUrl: req.ThumbnailUrl,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
	})
	if errors.Is(err, usecase.ErrBadReservationTimeRange) {
		return echo.NewHTTPError(http.StatusBadRequest, "bad reservation time range")
	}
	if unavailable, ok := errors.AsType[*usecase.ReservationSlotUnavailableError](err); ok {
		return echo.NewHTTPError(http.StatusBadRequest, unavailable.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newLivestream(livestream))
}
