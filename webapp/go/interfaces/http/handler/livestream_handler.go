package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type Livestream struct {
	ID           model.LivestreamID `json:"id"`
	Owner        User               `json:"owner"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	PlaylistUrl  string             `json:"playlist_url"`
	ThumbnailUrl string             `json:"thumbnail_url"`
	Tags         []Tag              `json:"tags"`
	StartAt      int64              `json:"start_at"`
	EndAt        int64              `json:"end_at"`
}

func newLivestream(l *model.Livestream) Livestream {
	tags := make([]Tag, len(l.Tags))
	for i, tag := range l.Tags {
		tags[i] = Tag{ID: tag.ID, Name: tag.Name}
	}
	return Livestream{
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

func newLivestreams(ls []*model.Livestream) []Livestream {
	livestreams := make([]Livestream, len(ls))
	for i, l := range ls {
		livestreams[i] = newLivestream(l)
	}
	return livestreams
}

type LivestreamHandler struct {
	livestreamUsecase usecase.LivestreamUsecase
}

func NewLivestreamHandler(livestreamUsecase usecase.LivestreamUsecase) *LivestreamHandler {
	return &LivestreamHandler{livestreamUsecase: livestreamUsecase}
}

// GET /api/livestream/:livestream_id
func (h *LivestreamHandler) GetLivestream(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	livestream, err := h.livestreamUsecase.FindByID(ctx, model.LivestreamID(livestreamID))
	if errors.Is(err, usecase.ErrLivestreamNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found livestream that has the given id")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestream(livestream))
}

// GET /api/livestream
func (h *LivestreamHandler) GetMyLivestreams(c echo.Context) error {
	ctx := c.Request().Context()
	if err := VerifyUserSession(c); err != nil {
		return err
	}

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
func (h *LivestreamHandler) GetUserLivestreams(c echo.Context) error {
	ctx := c.Request().Context()
	if err := VerifyUserSession(c); err != nil {
		return err
	}

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
func (h *LivestreamHandler) SearchLivestreams(c echo.Context) error {
	ctx := c.Request().Context()

	var livestreams []*model.Livestream
	var err error
	if tagName := c.QueryParam("tag"); tagName != "" {
		// タグによる取得 (limit は無視する)
		livestreams, err = h.livestreamUsecase.FindAllByTagName(ctx, tagName)
	} else {
		// 検索条件なし
		var limit *int64
		if c.QueryParam("limit") != "" {
			l, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "limit query parameter must be integer")
			}
			limit64 := int64(l)
			limit = &limit64
		}
		livestreams, err = h.livestreamUsecase.FindAll(ctx, limit)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, newLivestreams(livestreams))
}
