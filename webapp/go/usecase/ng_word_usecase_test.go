package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

func TestNGWordUsecase_FindAllByLivestreamID(t *testing.T) {
	want := []*model.NGWordModel{{ID: 1, Word: "bad"}}
	repo := &fakeNGWordRepository{ngWords: want}
	u := NewNGWordUsecase(&fakeTxManager{}, repo)

	got, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if repo.gotUserID != 1 || repo.gotLivestreamID != 10 {
		t.Errorf("userID = %d, livestreamID = %d", repo.gotUserID, repo.gotLivestreamID)
	}
}

func TestNGWordUsecase_FindAllByLivestreamID_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewNGWordUsecase(&fakeTxManager{}, &fakeNGWordRepository{err: boom})

	_, err := u.FindAllByLivestreamID(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
