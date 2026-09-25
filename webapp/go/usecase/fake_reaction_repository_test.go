package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeReactionRepository struct {
	reaction  *model.Reaction
	reactions []*model.Reaction
	err       error
	createID  model.ReactionID
	createErr error

	// calls は呼ばれたメソッド名を順に記録する
	calls           []string
	gotID           model.ReactionID
	gotLivestreamID model.LivestreamID
	gotLimit        model.Limit
	gotCreated      *model.ReactionModel

	// countsByOwnerID は CountByLivestreamOwnerID が返すリアクション数
	countsByOwnerID  map[model.UserID]int64
	countByOwnerErr  error
	countByOwnerName int64
	countByNameErr   error
	favoriteEmoji    string
	favoriteEmojiErr error
	gotOwnerNames    []string

	// countsByLivestreamID は CountByLivestreamID が返すリアクション数
	countsByLivestreamID map[model.LivestreamID]int64
	countByLivestreamErr error
	totalByLivestreamID  int64
	totalByLivestreamErr error
	gotTotalLivestreamID model.LivestreamID
}

func (r *fakeReactionRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.countsByLivestreamID[livestreamID], r.countByLivestreamErr
}

func (r *fakeReactionRepository) CountTotalByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	r.gotTotalLivestreamID = livestreamID
	return r.totalByLivestreamID, r.totalByLivestreamErr
}

func (r *fakeReactionRepository) CountByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID model.UserID) (int64, error) {
	return r.countsByOwnerID[userID], r.countByOwnerErr
}

func (r *fakeReactionRepository) CountByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	r.gotOwnerNames = append(r.gotOwnerNames, name)
	return r.countByOwnerName, r.countByNameErr
}

func (r *fakeReactionRepository) FindFavoriteEmojiByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (string, error) {
	r.gotOwnerNames = append(r.gotOwnerNames, name)
	return r.favoriteEmoji, r.favoriteEmojiErr
}

func (r *fakeReactionRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.ReactionID) (*model.Reaction, error) {
	r.calls = append(r.calls, "FindWithDetailsByID")
	r.gotID = id
	return r.reaction, r.err
}

func (r *fakeReactionRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Reaction, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamID")
	r.gotLivestreamID = livestreamID
	return r.reactions, r.err
}

func (r *fakeReactionRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit model.Limit) ([]*model.Reaction, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamIDLimited")
	r.gotLivestreamID = livestreamID
	r.gotLimit = limit
	return r.reactions, r.err
}

func (r *fakeReactionRepository) Create(ctx context.Context, q repository.Querier, reaction *model.ReactionModel) (model.ReactionID, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = reaction
	return r.createID, r.createErr
}
