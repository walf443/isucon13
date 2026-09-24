package model

import "testing"

func TestIconHash(t *testing.T) {
	tests := []struct {
		name  string
		image []byte
		want  string
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
			if got := IconHash(tt.image); got != tt.want {
				t.Errorf("IconHash(%q) = %s, want %s", tt.image, got, tt.want)
			}
		})
	}
}
