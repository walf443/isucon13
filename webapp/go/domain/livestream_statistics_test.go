package domain

import (
	"sort"
	"testing"
)

func TestLivestreamRanking_RankOf(t *testing.T) {
	ranking := LivestreamRanking{
		{LivestreamID: 1, Score: 10},
		{LivestreamID: 3, Score: 30},
		{LivestreamID: 2, Score: 10},
		{LivestreamID: 4, Score: 0},
	}
	sort.Sort(ranking)

	tests := []struct {
		livestreamID LivestreamID
		want         int64
	}{
		{livestreamID: 3, want: 1},
		// 同点の場合は ID の昇順に並ぶので、ID が大きい方が上位になる (移行前と同じ)
		{livestreamID: 2, want: 2},
		{livestreamID: 1, want: 3},
		{livestreamID: 4, want: 4},
		{livestreamID: 99, want: 5},
	}
	for _, tt := range tests {
		if got := ranking.RankOf(tt.livestreamID); got != tt.want {
			t.Errorf("RankOf(%d) = %d, want %d", tt.livestreamID, got, tt.want)
		}
	}
}
