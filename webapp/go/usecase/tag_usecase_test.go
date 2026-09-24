package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeTagRepository struct {
	tags []*model.TagModel
	err  error
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.TagModel, error) {
	return r.tags, r.err
}

func TestTagUsecase_FindAll(t *testing.T) {
	want := []*model.TagModel{{ID: 1, Name: "ライブ配信"}}
	u := NewTagUsecase(&fakeTxManager{}, &fakeTagRepository{tags: want})

	tags, err := u.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "ライブ配信" {
		t.Errorf("tags = %+v, want %+v", tags, want)
	}
}

func TestTagUsecase_FindAll_Error(t *testing.T) {
	wantErr := errors.New("boom")
	u := NewTagUsecase(&fakeTxManager{}, &fakeTagRepository{err: wantErr})

	_, err := u.FindAll(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}
