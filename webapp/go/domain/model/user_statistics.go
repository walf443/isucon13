package model

// UserStatistics は配信者としてのユーザの統計情報。
type UserStatistics struct {
	Rank              int64
	ViewersCount      int64
	TotalReactions    int64
	TotalLivecomments int64
	TotalTip          int64
	FavoriteEmoji     string
}

type UserRankingEntry struct {
	Username string
	Score    int64
}

// UserRanking はスコアの昇順 (同点ならユーザ名の昇順) に並ぶユーザのランキング。sort.Sort で並べ替えてから使う。
type UserRanking []UserRankingEntry

func (r UserRanking) Len() int      { return len(r) }
func (r UserRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r UserRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].Username < r[j].Username
	} else {
		return r[i].Score < r[j].Score
	}
}

// RankOf は並べ替え済みのランキングにおける username の順位 (末尾が 1 位) を返す。
// username がランキングに無い場合は len(r)+1 を返す。
func (r UserRanking) RankOf(username string) int64 {
	var rank int64 = 1
	for i := len(r) - 1; i >= 0; i-- {
		entry := r[i]
		if entry.Username == username {
			break
		}
		rank++
	}
	return rank
}
