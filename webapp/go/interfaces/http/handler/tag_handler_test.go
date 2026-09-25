package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type fakeTagUsecase struct {
	tags []*domain.TagModel
	err  error
}

func (s *fakeTagUsecase) FindAll(ctx context.Context) ([]*domain.TagModel, error) {
	return s.tags, s.err
}

func TestTagHandler_GetTags(t *testing.T) {
	tests := []struct {
		name     string
		usecase  *fakeTagUsecase
		wantCode int
		wantBody string
	}{
		{
			name: "returns tags",
			usecase: &fakeTagUsecase{tags: []*domain.TagModel{
				{ID: 1, Name: "ライブ配信"},
				{ID: 2, Name: "ゲーム実況"},
			}},
			wantCode: http.StatusOK,
			wantBody: `{"tags":[{"id":1,"name":"ライブ配信"},{"id":2,"name":"ゲーム実況"}]}` + "\n",
		},
		{
			name:     "returns empty array when no tags",
			usecase:  &fakeTagUsecase{tags: nil},
			wantCode: http.StatusOK,
			wantBody: `{"tags":[]}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			usecase:  &fakeTagUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newTagHandler(tt.usecase).GetTags, testRequest{method: http.MethodGet, route: "/api/tag", path: "/api/tag"})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
		})
	}
}
