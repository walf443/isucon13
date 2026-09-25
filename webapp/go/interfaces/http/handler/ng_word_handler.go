package handler

import (
	"net/http"
	"strconv"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type NGWord struct {
	ID           model.NGWordID     `json:"id"`
	UserID       model.UserID       `json:"user_id"`
	LivestreamID model.LivestreamID `json:"livestream_id"`
	Word         string             `json:"word"`
	CreatedAt    int64              `json:"created_at"`
}

type NGWordHandler struct {
	ngWordUsecase usecase.NGWordUsecase
}

func NewNGWordHandler(ngWordUsecase usecase.NGWordUsecase) *NGWordHandler {
	return &NGWordHandler{ngWordUsecase: ngWordUsecase}
}

// GET /api/livestream/:livestream_id/ngwords
func (h *NGWordHandler) GetNGWords(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
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

	ngWordModels, err := h.ngWordUsecase.FindAllByLivestreamID(ctx, userID, model.LivestreamID(livestreamID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// NG ワードが無い場合は [] ではなく null を返す (移行前と同じ)
	var ngWords []NGWord
	for _, w := range ngWordModels {
		ngWords = append(ngWords, NGWord{
			ID:           w.ID,
			UserID:       w.UserID,
			LivestreamID: w.LivestreamID,
			Word:         w.Word,
			CreatedAt:    w.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, ngWords)
}
