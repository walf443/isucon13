package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeLivestreamRepository struct {
	livestreamModel  *model.LivestreamModel
	livestreamModels []*model.LivestreamModel
	livestream       *model.Livestream
	livestreams      []*model.Livestream
	err              error
	createID         model.LivestreamID
	createErr        error
	addTagErr        error
	findAllErr       error

	gotID      model.LivestreamID
	gotUserID  model.UserID
	gotTagIDs  []model.TagID
	gotLimit   model.Limit
	gotCreated *model.LivestreamModel
	// addedTagIDs は AddTag に渡されたタグ ID
	addedTagIDs []model.TagID
	// calls は呼ばれたメソッド名を順に記録する
	calls []string
}

func (r *fakeLivestreamRepository) Create(ctx context.Context, q repository.Querier, livestream *model.LivestreamModel) (model.LivestreamID, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = livestream
	return r.createID, r.createErr
}

func (r *fakeLivestreamRepository) AddTag(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, tagID model.TagID) error {
	r.calls = append(r.calls, "AddTag")
	r.gotID = livestreamID
	r.addedTagIDs = append(r.addedTagIDs, tagID)
	return r.addTagErr
}

func (r *fakeLivestreamRepository) FindByID(ctx context.Context, q repository.Querier, id model.LivestreamID) (*model.LivestreamModel, error) {
	r.calls = append(r.calls, "FindByID")
	r.gotID = id
	return r.livestreamModel, r.err
}

func (r *fakeLivestreamRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.LivestreamModel, error) {
	r.calls = append(r.calls, "FindAll")
	return r.livestreamModels, r.findAllErr
}

func (r *fakeLivestreamRepository) FindAllByUserID(ctx context.Context, q repository.Querier, userID model.UserID) ([]*model.LivestreamModel, error) {
	r.calls = append(r.calls, "FindAllByUserID")
	r.gotUserID = userID
	return r.livestreamModels, r.err
}

func (r *fakeLivestreamRepository) FindAllByIDAndUserID(ctx context.Context, q repository.Querier, id model.LivestreamID, userID model.UserID) ([]*model.LivestreamModel, error) {
	r.calls = append(r.calls, "FindAllByIDAndUserID")
	r.gotID = id
	r.gotUserID = userID
	return r.livestreamModels, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetailsByTagIDs(ctx context.Context, q repository.Querier, tagIDs []model.TagID) ([]*model.Livestream, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByTagIDs")
	r.gotTagIDs = tagIDs
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetails(ctx context.Context, q repository.Querier) ([]*model.Livestream, error) {
	r.calls = append(r.calls, "FindAllWithDetails")
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetailsLimited(ctx context.Context, q repository.Querier, limit model.Limit) ([]*model.Livestream, error) {
	r.calls = append(r.calls, "FindAllWithDetailsLimited")
	r.gotLimit = limit
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetailsByUserID(ctx context.Context, q repository.Querier, userID model.UserID) ([]*model.Livestream, error) {
	r.gotUserID = userID
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivestreamID) (*model.Livestream, error) {
	r.calls = append(r.calls, "FindWithDetailsByID")
	r.gotID = id
	return r.livestream, r.err
}
