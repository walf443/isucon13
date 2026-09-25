package domain

import (
	"sort"
	"testing"
)

func TestUserRanking_RankOf(t *testing.T) {
	ranking := UserRanking{
		{Username: "alice", Score: 10},
		{Username: "carol", Score: 30},
		{Username: "bob", Score: 10},
		{Username: "dave", Score: 0},
	}
	sort.Sort(ranking)

	tests := []struct {
		username string
		want     int64
	}{
		{username: "carol", want: 1},
		// 同点の場合はユーザ名の昇順に並ぶので、名前が後ろの方が上位になる (移行前と同じ)
		{username: "bob", want: 2},
		{username: "alice", want: 3},
		{username: "dave", want: 4},
		{username: "unknown", want: 5},
	}
	for _, tt := range tests {
		if got := ranking.RankOf(tt.username); got != tt.want {
			t.Errorf("RankOf(%q) = %d, want %d", tt.username, got, tt.want)
		}
	}
}
