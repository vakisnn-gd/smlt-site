package main

import "testing"

func TestNormalizeYouTubeID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		valid bool
	}{
		{name: "raw id", input: "dQw4w9WgXcQ", want: "dQw4w9WgXcQ", valid: true},
		{name: "watch url", input: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=42s", want: "dQw4w9WgXcQ", valid: true},
		{name: "short url", input: "https://youtu.be/dQw4w9WgXcQ?si=test", want: "dQw4w9WgXcQ", valid: true},
		{name: "shorts url", input: "youtube.com/shorts/dQw4w9WgXcQ", want: "dQw4w9WgXcQ", valid: true},
		{name: "embed url", input: "https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ", want: "dQw4w9WgXcQ", valid: true},
		{name: "other host", input: "https://example.com/watch?v=dQw4w9WgXcQ", valid: false},
		{name: "bad id", input: "dQw4w9WgXc", valid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := normalizeYouTubeID(tt.input)
			if got != tt.want || valid != tt.valid {
				t.Fatalf("normalizeYouTubeID(%q) = %q, %v; want %q, %v", tt.input, got, valid, tt.want, tt.valid)
			}
		})
	}
}
