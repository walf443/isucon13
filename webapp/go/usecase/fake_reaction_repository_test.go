package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeReactionRepository はテストで設定した関数に処理を委ねる ReactionRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeReactionRepository struct {
	repository.ReactionRepository

	findWithDetailsByID                     func(ctx context.Context, q repository.Querier, id domain.ReactionID) (*domain.Reaction, error)
	findAllWithDetailsByLivestreamID        func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Reaction, error)
	findAllWithDetailsByLivestreamIDLimited func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Reaction, error)
	create                                  func(ctx context.Context, q repository.Querier, reaction *domain.ReactionModel) (domain.ReactionID, error)
	countByLivestreamID                     func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error)
	countTotalByLivestreamID                func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error)
	countByLivestreamOwnerID                func(ctx context.Context, q repository.Querier, userID domain.UserID) (int64, error)
	countByLivestreamOwnerName              func(ctx context.Context, q repository.Querier, name string) (int64, error)
	findFavoriteEmojiByLivestreamOwnerName  func(ctx context.Context, q repository.Querier, name string) (string, error)
}

func (r *fakeReactionRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id domain.ReactionID) (*domain.Reaction, error) {
	return r.findWithDetailsByID(ctx, q, id)
}

func (r *fakeReactionRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.Reaction, error) {
	return r.findAllWithDetailsByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeReactionRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, limit domain.Limit) ([]*domain.Reaction, error) {
	return r.findAllWithDetailsByLivestreamIDLimited(ctx, q, livestreamID, limit)
}

func (r *fakeReactionRepository) Create(ctx context.Context, q repository.Querier, reaction *domain.ReactionModel) (domain.ReactionID, error) {
	return r.create(ctx, q, reaction)
}

func (r *fakeReactionRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	return r.countByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeReactionRepository) CountTotalByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) (int64, error) {
	return r.countTotalByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeReactionRepository) CountByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID domain.UserID) (int64, error) {
	return r.countByLivestreamOwnerID(ctx, q, userID)
}

func (r *fakeReactionRepository) CountByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	return r.countByLivestreamOwnerName(ctx, q, name)
}

func (r *fakeReactionRepository) FindFavoriteEmojiByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (string, error) {
	return r.findFavoriteEmojiByLivestreamOwnerName(ctx, q, name)
}
