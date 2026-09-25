package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeTxManager struct{}

func (m *fakeTxManager) RunInTx(ctx context.Context, fn func(q repository.Querier) error) error {
	return fn(nil)
}

type fakeTagRepository struct {
	tags []*model.TagModel
	ids  []int64
	err  error

	gotName string
}

func (r *fakeTagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]int64, error) {
	r.gotName = name
	return r.ids, r.err
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.TagModel, error) {
	return r.tags, r.err
}

type fakeUserRepository struct {
	id          int64
	user        *model.UserModel
	userDetails *model.User
	err         error

	gotID   int64
	gotName string
}

func (r *fakeUserRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	r.gotName = name
	return r.id, r.err
}

func (r *fakeUserRepository) FindByID(ctx context.Context, q repository.Querier, id int64) (*model.UserModel, error) {
	r.gotID = id
	return r.user, r.err
}

func (r *fakeUserRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*model.UserModel, error) {
	r.gotName = name
	return r.user, r.err
}

func (r *fakeUserRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id int64) (*model.User, error) {
	r.gotID = id
	return r.userDetails, r.err
}

func (r *fakeUserRepository) FindWithDetailsByName(ctx context.Context, q repository.Querier, name string) (*model.User, error) {
	r.gotName = name
	return r.userDetails, r.err
}

type fakeThemeRepository struct {
	theme     *model.ThemeModel
	err       error
	gotUserID int64
}

func (r *fakeThemeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID int64) (*model.ThemeModel, error) {
	r.gotUserID = userID
	return r.theme, r.err
}

type fakeIconRepository struct {
	image     []byte
	err       error
	createID  int64
	createErr error
	deleteErr error

	// calls は呼ばれたメソッド名を順に記録する
	calls     []string
	gotUserID int64
	gotImage  []byte
}

func (r *fakeIconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID int64) ([]byte, error) {
	r.calls = append(r.calls, "FindImageByUserID")
	r.gotUserID = userID
	return r.image, r.err
}

func (r *fakeIconRepository) Create(ctx context.Context, q repository.Querier, userID int64, image []byte) (int64, error) {
	r.calls = append(r.calls, "Create")
	r.gotUserID = userID
	r.gotImage = image
	return r.createID, r.createErr
}

func (r *fakeIconRepository) DeleteByUserID(ctx context.Context, q repository.Querier, userID int64) error {
	r.calls = append(r.calls, "DeleteByUserID")
	r.gotUserID = userID
	return r.deleteErr
}

type fakeLivestreamRepository struct {
	livestream  *model.Livestream
	livestreams []*model.Livestream
	err         error

	gotID     int64
	gotUserID int64
	gotTagIDs []int64
	gotLimit  int64
	// calls は呼ばれたメソッド名を順に記録する
	calls []string
}

func (r *fakeLivestreamRepository) FindAllWithDetailsByTagIDs(ctx context.Context, q repository.Querier, tagIDs []int64) ([]*model.Livestream, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByTagIDs")
	r.gotTagIDs = tagIDs
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetails(ctx context.Context, q repository.Querier) ([]*model.Livestream, error) {
	r.calls = append(r.calls, "FindAllWithDetails")
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetailsLimited(ctx context.Context, q repository.Querier, limit int64) ([]*model.Livestream, error) {
	r.calls = append(r.calls, "FindAllWithDetailsLimited")
	r.gotLimit = limit
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindAllWithDetailsByUserID(ctx context.Context, q repository.Querier, userID int64) ([]*model.Livestream, error) {
	r.gotUserID = userID
	return r.livestreams, r.err
}

func (r *fakeLivestreamRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id int64) (*model.Livestream, error) {
	r.gotID = id
	return r.livestream, r.err
}

type fakeReactionRepository struct {
	reaction  *model.Reaction
	reactions []*model.Reaction
	err       error
	createID  int64
	createErr error

	// calls は呼ばれたメソッド名を順に記録する
	calls           []string
	gotID           int64
	gotLivestreamID int64
	gotLimit        int64
	gotCreated      *model.ReactionModel
}

func (r *fakeReactionRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id int64) (*model.Reaction, error) {
	r.calls = append(r.calls, "FindWithDetailsByID")
	r.gotID = id
	return r.reaction, r.err
}

func (r *fakeReactionRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID int64) ([]*model.Reaction, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamID")
	r.gotLivestreamID = livestreamID
	return r.reactions, r.err
}

func (r *fakeReactionRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID int64, limit int64) ([]*model.Reaction, error) {
	r.calls = append(r.calls, "FindAllWithDetailsByLivestreamIDLimited")
	r.gotLivestreamID = livestreamID
	r.gotLimit = limit
	return r.reactions, r.err
}

func (r *fakeReactionRepository) Create(ctx context.Context, q repository.Querier, reaction *model.ReactionModel) (int64, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = reaction
	return r.createID, r.createErr
}
