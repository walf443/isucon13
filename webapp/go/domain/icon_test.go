package domain

import "testing"

func TestHashIcon(t *testing.T) {
	tests := []struct {
		name  string
		image []byte
		want  IconHash
	}{
		{
			name:  "empty image",
			image: []byte{},
			want:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:  "some image",
			image: []byte("icon"),
			want:  "c2d4b446a44ce54fab8e01150e24dd24f3d850c7c14dcfe31f6321341dd86874",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashIcon(tt.image); got != tt.want {
				t.Errorf("HashIcon(%q) = %s, want %s", tt.image, got, tt.want)
			}
		})
	}
}

func TestUserIconHash(t *testing.T) {
	defaultIconHash := HashIcon([]byte("default"))

	tests := []struct {
		name       string
		image      []byte
		registered bool
		want       IconHash
	}{
		{name: "registered", image: []byte("icon"), registered: true, want: HashIcon([]byte("icon"))},
		{name: "not registered", image: nil, registered: false, want: defaultIconHash},
		// 空の画像を登録している場合は既定のアイコンにはしない
		{name: "registered empty image", image: []byte{}, registered: true, want: HashIcon([]byte{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserIconHash(tt.image, tt.registered, defaultIconHash); got != tt.want {
				t.Errorf("UserIconHash() = %s, want %s", got, tt.want)
			}
		})
	}
}
