package domain

import "testing"

func TestLivestreamModel_IsOwnedBy(t *testing.T) {
	l := &LivestreamModel{ID: 10, UserID: 1}

	tests := []struct {
		name   string
		userID UserID
		want   bool
	}{
		{name: "owner", userID: 1, want: true},
		{name: "other user", userID: 2, want: false},
		{
			// ライブ配信の ID ではなく配信者の ID と比較していることを確認する
			name:   "user whose id equals the livestream id",
			userID: 10,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := l.IsOwnedBy(tt.userID); got != tt.want {
				t.Errorf("IsOwnedBy(%d) = %v, want %v", tt.userID, got, tt.want)
			}
		})
	}
}
