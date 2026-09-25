package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type ngWordResponse struct {
	ID           model.NGWordID     `json:"id"`
	UserID       model.UserID       `json:"user_id"`
	LivestreamID model.LivestreamID `json:"livestream_id"`
	Word         string             `json:"word"`
	CreatedAt    int64              `json:"created_at"`
}

type moderateRequest struct {
	NGWord string `json:"ng_word"`
}

type ngWordHandler struct {
	ngWordUsecase usecase.NGWordUsecase
}

func newNGWordHandler(ngWordUsecase usecase.NGWordUsecase) *ngWordHandler {
	return &ngWordHandler{ngWordUsecase: ngWordUsecase}
}

// GET /api/livestream/:livestream_id/ngwords
func (h *ngWordHandler) GetNGWords(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	ngWordModels, err := h.ngWordUsecase.FindAllByLivestreamID(ctx, userID, livestreamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// NG ワードが無い場合は [] ではなく null を返す (移行前と同じ)
	var ngWords []ngWordResponse
	for _, w := range ngWordModels {
		ngWords = append(ngWords, ngWordResponse{
			ID:           w.ID,
			UserID:       w.UserID,
			LivestreamID: w.LivestreamID,
			Word:         w.Word,
			CreatedAt:    w.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, ngWords)
}

// 配信者によるモデレーション (NGワード登録)
// POST /api/livestream/:livestream_id/moderate
func (h *ngWordHandler) Moderate(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	var req moderateRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	wordID, err := h.ngWordUsecase.Moderate(ctx, userID, livestreamID, req.NGWord)
	if errors.Is(err, usecase.ErrNotLivestreamOwner) {
		return echo.NewHTTPError(http.StatusBadRequest, "A streamer can't moderate livestreams that other streamers own")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"word_id": wordID,
	})
}
