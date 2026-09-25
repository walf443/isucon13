package handler

import (
	"encoding/json"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type reactionResponse struct {
	ID         model.ReactionID   `json:"id"`
	EmojiName  string             `json:"emoji_name"`
	User       userResponse       `json:"user"`
	Livestream livestreamResponse `json:"livestream"`
	CreatedAt  int64              `json:"created_at"`
}

func newReaction(r *model.Reaction) reactionResponse {
	return reactionResponse{
		ID:         r.ID,
		EmojiName:  r.EmojiName,
		User:       newUser(&r.User),
		Livestream: newLivestream(&r.Livestream),
		CreatedAt:  r.CreatedAt,
	}
}

type postReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

type reactionHandler struct {
	reactionUsecase usecase.ReactionUsecase
}

func newReactionHandler(reactionUsecase usecase.ReactionUsecase) *reactionHandler {
	return &reactionHandler{reactionUsecase: reactionUsecase}
}

// GET /api/livestream/:livestream_id/reaction
func (h *reactionHandler) GetReactions(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	limit, err := parseLimitQueryParam(c, maxReactionsLimit)
	if err != nil {
		return err
	}

	reactionModels, err := h.reactionUsecase.FindAllByLivestreamID(ctx, livestreamID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactions := make([]reactionResponse, len(reactionModels))
	for i, r := range reactionModels {
		reactions[i] = newReaction(r)
	}
	return c.JSON(http.StatusOK, reactions)
}

// POST /api/livestream/:livestream_id/reaction
func (h *reactionHandler) PostReaction(c echo.Context) error {
	ctx := c.Request().Context()
	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	var req postReactionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	reaction, err := h.reactionUsecase.Create(ctx, userID, livestreamID, req.EmojiName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newReaction(reaction))
}
