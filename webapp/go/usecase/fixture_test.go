// Filler と、Filler を使う usecase のテストで共有するテスト用のデータ。
//
// detailFixture が持つユーザ・タグから fake の repository と本物の Filler を作り、
// 同じデータから期待する XxxDetail を作る (user / livestream / reaction / livecomment / report)。
// testXxx の変数は、Filler に渡す行 (domain のエンティティ) のテスト用のデータ。
package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// detailFixture は Filler が引くデータ (ユーザ・タグ) を持つ。
// ユーザのテーマは ID がユーザの ID の 10 倍のものとし、アイコンは未登録とする。
type detailFixture struct {
	users map[domain.UserID]*domain.User
	tags  map[domain.TagID]*domain.Tag
	// livestreamTags はライブ配信ごとに付いているタグの ID。
	livestreamTags map[domain.LivestreamID][]domain.TagID
	// calls は呼ばれた repository のメソッドと引数を記録する。
	calls []string
}

// testDetailFixture は配信者 alice (ID 42)・視聴者 bob (ID 43) とタグ 7, 8 を持ち、
// ライブ配信 1 にタグ 7, 8、ライブ配信 2 にタグ 7 が付いている。
func testDetailFixture() *detailFixture {
	return &detailFixture{
		users: map[domain.UserID]*domain.User{
			42: {ID: 42, Name: "alice"},
			43: {ID: 43, Name: "bob", DisplayName: "Bob"},
		},
		tags: map[domain.TagID]*domain.Tag{
			7: {ID: 7, Name: "ゲーム実況"},
			8: {ID: 8, Name: "雑談"},
		},
		livestreamTags: map[domain.LivestreamID][]domain.TagID{1: {7, 8}, 2: {7}},
	}
}

func (f *detailFixture) record(call string) {
	f.calls = append(f.calls, call)
}

// userRepo は f のユーザを ID の一覧でまとめて引ける fakeUserRepository を返す。
func (f *detailFixture) userRepo() *fakeUserRepository {
	return &fakeUserRepository{
		findAllByIDs: func(_ context.Context, _ repository.Querier, ids []domain.UserID) ([]*domain.User, error) {
			f.record(fmt.Sprintf("users %v", ids))
			var found []*domain.User
			for _, id := range ids {
				if user, ok := f.users[id]; ok {
					found = append(found, user)
				}
			}
			return found, nil
		},
	}
}

// userFiller は f のユーザのテーマを返す UserFiller を返す。
func (f *detailFixture) userFiller() *UserFiller {
	themes := map[domain.UserID]*domain.Theme{}
	for id := range f.users {
		themes[id] = &domain.Theme{ID: domain.ThemeID(id * 10), UserID: id}
	}
	return newUserFillerForTest(themes, nil, nil)
}

// user は、このデータでユーザ id を埋めた結果として期待する domain.UserDetail を返す。
func (f *detailFixture) user(id domain.UserID) domain.UserDetail {
	user := f.users[id]
	return domain.UserDetail{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Theme:       domain.Theme{ID: domain.ThemeID(user.ID * 10), UserID: user.ID},
		IconHash:    "default-hash",
	}
}

func (f *detailFixture) livestreamFiller() *LivestreamFiller {
	livestreamTagRepo := &fakeLivestreamTagRepository{
		findAllByLivestreamIDs: func(_ context.Context, _ repository.Querier, livestreamIDs []domain.LivestreamID) ([]*domain.LivestreamTag, error) {
			f.record(fmt.Sprintf("livestream tags %v", livestreamIDs))
			var livestreamTags []*domain.LivestreamTag
			for _, livestreamID := range livestreamIDs {
				for _, tagID := range f.livestreamTags[livestreamID] {
					livestreamTags = append(livestreamTags, &domain.LivestreamTag{LivestreamID: livestreamID, TagID: tagID})
				}
			}
			return livestreamTags, nil
		},
	}
	tagRepo := &fakeTagRepository{
		findAllByIDs: func(_ context.Context, _ repository.Querier, ids []domain.TagID) ([]*domain.Tag, error) {
			f.record(fmt.Sprintf("tags %v", ids))
			var found []*domain.Tag
			for _, id := range ids {
				if tag, ok := f.tags[id]; ok {
					found = append(found, tag)
				}
			}
			return found, nil
		},
	}
	return NewLivestreamFiller(f.userRepo(), livestreamTagRepo, tagRepo, f.userFiller())
}

// livestream は、このデータで livestream を埋めた結果として期待する domain.LivestreamDetail を返す。
func (f *detailFixture) livestream(livestream *domain.Livestream) *domain.LivestreamDetail {
	tags := make([]domain.Tag, len(f.livestreamTags[livestream.ID]))
	for i, tagID := range f.livestreamTags[livestream.ID] {
		tags[i] = *f.tags[tagID]
	}
	return &domain.LivestreamDetail{
		ID:           livestream.ID,
		Owner:        f.user(livestream.UserID),
		Title:        livestream.Title,
		Description:  livestream.Description,
		PlaylistUrl:  livestream.PlaylistUrl,
		ThumbnailUrl: livestream.ThumbnailUrl,
		Tags:         tags,
		StartAt:      livestream.StartAt,
		EndAt:        livestream.EndAt,
	}
}

var (
	testLivestream1 = &domain.Livestream{ID: 1, UserID: 42, Title: "first", Description: "desc", PlaylistUrl: "p", ThumbnailUrl: "t", StartAt: 100, EndAt: 200}
	testLivestream2 = &domain.Livestream{ID: 2, UserID: 42, Title: "second"}
)

// newUserFillerForTest は themes / icons をユーザの ID ごとに返す UserFiller を返す。
// themes に無いユーザはテーマ欠損、icons に無いユーザはアイコン未登録として扱う。
// テーマをまとめて引いたときのユーザの ID の一覧を themeCalls に記録する。
func newUserFillerForTest(themes map[domain.UserID]*domain.Theme, icons map[domain.UserID][]byte, themeCalls *[][]domain.UserID) *UserFiller {
	themeRepo := &fakeThemeRepository{
		findAllByUserIDs: func(_ context.Context, _ repository.Querier, userIDs []domain.UserID) ([]*domain.Theme, error) {
			if themeCalls != nil {
				*themeCalls = append(*themeCalls, userIDs)
			}
			var found []*domain.Theme
			for _, userID := range userIDs {
				if theme, ok := themes[userID]; ok {
					found = append(found, theme)
				}
			}
			return found, nil
		},
	}
	iconRepo := &fakeIconRepository{
		findAllByUserIDs: func(_ context.Context, _ repository.Querier, userIDs []domain.UserID) ([]*domain.Icon, error) {
			var found []*domain.Icon
			for _, userID := range userIDs {
				if image, ok := icons[userID]; ok {
					found = append(found, &domain.Icon{UserID: userID, Image: image})
				}
			}
			return found, nil
		},
	}
	return NewUserFiller(themeRepo, iconRepo, "default-hash")
}

// newLivestreamRepositoryWithLivestreams は ID で models を引ける fakeLivestreamRepository を返す。
// FindByID は見つからない ID に ErrNotFound を返し、FindAllByIDs は見つからない ID を結果に含めない。
// FindAllByIDs で引いた ID の一覧を calls に記録する。
func newLivestreamRepositoryWithLivestreams(calls *[][]domain.LivestreamID, models ...*domain.Livestream) *fakeLivestreamRepository {
	byID := indexBy(models, func(m *domain.Livestream) domain.LivestreamID { return m.ID })
	return &fakeLivestreamRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivestreamID) (*domain.Livestream, error) {
			m, ok := byID[id]
			if !ok {
				return nil, repository.ErrNotFound
			}
			return m, nil
		},
		findAllByIDs: func(_ context.Context, _ repository.Querier, ids []domain.LivestreamID) ([]*domain.Livestream, error) {
			if calls != nil {
				*calls = append(*calls, ids)
			}
			var found []*domain.Livestream
			for _, id := range ids {
				if m, ok := byID[id]; ok {
					found = append(found, m)
				}
			}
			return found, nil
		},
	}
}

// reactionFiller は f のユーザと livestreams のライブ配信を引く ReactionFiller を返す。
func (f *detailFixture) reactionFiller(livestreams ...*domain.Livestream) *ReactionFiller {
	return NewReactionFiller(f.userRepo(), newLivestreamRepositoryWithLivestreams(nil, livestreams...), f.userFiller(), f.livestreamFiller())
}

// reaction は、このデータで reaction を埋めた結果として期待する domain.ReactionDetail を返す。
// livestream は reaction のライブ配信。
func (f *detailFixture) reaction(reaction *domain.Reaction, livestream *domain.Livestream) *domain.ReactionDetail {
	return &domain.ReactionDetail{
		ID:         reaction.ID,
		EmojiName:  reaction.EmojiName,
		User:       f.user(reaction.UserID),
		Livestream: *f.livestream(livestream),
		CreatedAt:  reaction.CreatedAt,
	}
}

var (
	testReaction1 = &domain.Reaction{ID: 1, UserID: 43, LivestreamID: 1, EmojiName: "tada", CreatedAt: 300}
	testReaction2 = &domain.Reaction{ID: 2, UserID: 42, LivestreamID: 1, EmojiName: "heart", CreatedAt: 200}
	testReaction3 = &domain.Reaction{ID: 3, UserID: 43, LivestreamID: 2, EmojiName: "tada", CreatedAt: 100}
)

// newLivecommentRepositoryWithLivecomments は ID で models を引ける fakeLivecommentRepository を返す。
// FindByID は見つからない ID に ErrNotFound を返し、FindAllByIDs は見つからない ID を結果に含めない。
func newLivecommentRepositoryWithLivecomments(models ...*domain.Livecomment) *fakeLivecommentRepository {
	byID := indexBy(models, func(m *domain.Livecomment) domain.LivecommentID { return m.ID })
	return &fakeLivecommentRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentID) (*domain.Livecomment, error) {
			m, ok := byID[id]
			if !ok {
				return nil, repository.ErrNotFound
			}
			return m, nil
		},
		findAllByIDs: func(_ context.Context, _ repository.Querier, ids []domain.LivecommentID) ([]*domain.Livecomment, error) {
			var found []*domain.Livecomment
			for _, id := range ids {
				if m, ok := byID[id]; ok {
					found = append(found, m)
				}
			}
			return found, nil
		},
	}
}

// livecommentFiller は f のユーザと livestreams のライブ配信を引く LivecommentFiller を返す。
func (f *detailFixture) livecommentFiller(livestreams ...*domain.Livestream) *LivecommentFiller {
	return NewLivecommentFiller(f.userRepo(), newLivestreamRepositoryWithLivestreams(nil, livestreams...), f.userFiller(), f.livestreamFiller())
}

// livecomment は、このデータで livecomment を埋めた結果として期待する domain.LivecommentDetail を返す。
// livestream は livecomment のライブ配信。
func (f *detailFixture) livecomment(livecomment *domain.Livecomment, livestream *domain.Livestream) *domain.LivecommentDetail {
	return &domain.LivecommentDetail{
		ID:         livecomment.ID,
		User:       f.user(livecomment.UserID),
		Livestream: *f.livestream(livestream),
		Comment:    livecomment.Comment,
		Tip:        livecomment.Tip,
		CreatedAt:  livecomment.CreatedAt,
	}
}

var (
	testLivecomment1 = &domain.Livecomment{ID: 50, UserID: 43, LivestreamID: 1, Comment: "hello", Tip: 10, CreatedAt: 300}
	testLivecomment2 = &domain.Livecomment{ID: 51, UserID: 42, LivestreamID: 1, Comment: "thanks", CreatedAt: 200}
	testLivecomment3 = &domain.Livecomment{ID: 52, UserID: 43, LivestreamID: 2, Comment: "hi", CreatedAt: 100}
)

// reportFiller は f のユーザ、livecomments のライブコメント、testLivestream1・2 のライブ配信を引く LivecommentReportFiller を返す。
func (f *detailFixture) reportFiller(livecomments ...*domain.Livecomment) *LivecommentReportFiller {
	return NewLivecommentReportFiller(f.userRepo(), newLivecommentRepositoryWithLivecomments(livecomments...), f.userFiller(), f.livecommentFiller(testLivestream1, testLivestream2))
}

// report は、このデータで report を埋めた結果として期待する domain.LivecommentReportDetail を返す。
// livecomment, livestream は報告されたライブコメントとそのライブ配信。
func (f *detailFixture) report(report *domain.LivecommentReport, livecomment *domain.Livecomment, livestream *domain.Livestream) *domain.LivecommentReportDetail {
	return &domain.LivecommentReportDetail{
		ID:          report.ID,
		Reporter:    f.user(report.UserID),
		Livecomment: *f.livecomment(livecomment, livestream),
		CreatedAt:   report.CreatedAt,
	}
}

var (
	testReport1 = &domain.LivecommentReport{ID: 7, UserID: 42, LivestreamID: 1, LivecommentID: 50, CreatedAt: 400}
	testReport2 = &domain.LivecommentReport{ID: 8, UserID: 42, LivestreamID: 1, LivecommentID: 51, CreatedAt: 500}
)
