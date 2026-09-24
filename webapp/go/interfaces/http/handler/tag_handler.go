package handler

import (
	"net/http"

	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

type TagHandler struct {
	tagUsecase usecase.TagUsecase
}

func NewTagHandler(tagUsecase usecase.TagUsecase) *TagHandler {
	return &TagHandler{tagUsecase: tagUsecase}
}

// GET /api/tag
func (h *TagHandler) GetTags(c echo.Context) error {
	ctx := c.Request().Context()

	tagModels, err := h.tagUsecase.FindAll(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get tags: "+err.Error())
	}

	tags := make([]*Tag, len(tagModels))
	for i := range tagModels {
		tags[i] = &Tag{
			ID:   tagModels[i].ID,
			Name: tagModels[i].Name,
		}
	}
	return c.JSON(http.StatusOK, &TagsResponse{
		Tags: tags,
	})
}
