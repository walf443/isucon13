package domain

// LivestreamStatistics はライブ配信の統計情報。
type LivestreamStatistics struct {
	Rank           int64
	ViewersCount   int64
	TotalReactions int64
	TotalReports   int64
	MaxTip         int64
}

type LivestreamRankingEntry struct {
	LivestreamID LivestreamID
	Score        int64
}

// LivestreamRanking はスコアの昇順 (同点ならライブ配信 ID の昇順) に並ぶライブ配信のランキング。sort.Sort で並べ替えてから使う。
type LivestreamRanking []LivestreamRankingEntry

func (r LivestreamRanking) Len() int      { return len(r) }
func (r LivestreamRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r LivestreamRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].LivestreamID < r[j].LivestreamID
	} else {
		return r[i].Score < r[j].Score
	}
}

// RankOf は並べ替え済みのランキングにおける livestreamID の順位 (末尾が 1 位) を返す。
// livestreamID がランキングに無い場合は len(r)+1 を返す。
func (r LivestreamRanking) RankOf(livestreamID LivestreamID) int64 {
	var rank int64 = 1
	for i := len(r) - 1; i >= 0; i-- {
		entry := r[i]
		if entry.LivestreamID == livestreamID {
			break
		}
		rank++
	}
	return rank
}
