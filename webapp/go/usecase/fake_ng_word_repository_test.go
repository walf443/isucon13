package usecase

import (
	"context"
	"slices"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeNGWordRepository struct {
	ngWords   []*model.NGWordModel
	err       error
	createID  model.NGWordID
	createErr error

	// calls は呼ばれたメソッド名を順に記録する
	calls      []string
	gotCreated *model.NGWordModel
	// hitWords は Matches で当たりとする NG ワード
	hitWords []string
	matchErr error

	gotUserID       model.UserID
	gotLivestreamID model.LivestreamID
	gotComments     []string
	matchedWords    []string
}

func (r *fakeNGWordRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	r.calls = append(r.calls, "FindAllByLivestreamID")
	r.gotLivestreamID = livestreamID
	return r.ngWords, r.err
}

func (r *fakeNGWordRepository) Create(ctx context.Context, q repository.Querier, ngWord *model.NGWordModel) (model.NGWordID, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = ngWord
	return r.createID, r.createErr
}

func (r *fakeNGWordRepository) Matches(ctx context.Context, q repository.Querier, comment string, word string) (bool, error) {
	r.gotComments = append(r.gotComments, comment)
	r.matchedWords = append(r.matchedWords, word)
	if r.matchErr != nil {
		return false, r.matchErr
	}
	return slices.Contains(r.hitWords, word), nil
}

func (r *fakeNGWordRepository) FindAllByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	r.gotUserID = userID
	r.gotLivestreamID = livestreamID
	return r.ngWords, r.err
}
