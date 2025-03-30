package main

import (
	"regexp"
	"testing"
)

func TestScreenshotRegex(t *testing.T) {
	re := regexp.MustCompile(screenshotPattern)

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "valid PM",
			filename: "Screen Shot 2020-06-21 at 4.21.35 PM.png",
			want:     true,
		},
		{
			name:     "valid AM",
			filename: "Screen Shot 2020-06-21 at 4.21.35 AM.png",
			want:     true,
		},
		{
			name:     "valid PM 2",
			filename: "Screenshot 2025-03-29 at 11.16.20 PM.png",
			want:     true,
		},
		{
			name:     "valid PM with narrow no-break space",
			filename: "Screenshot 2025-03-30 at 2.06.46 PM.png",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := re.MatchString(tt.filename); got != tt.want {
				t.Errorf("regex.MatchString(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
