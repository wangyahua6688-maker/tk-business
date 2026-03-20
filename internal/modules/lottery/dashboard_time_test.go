package lottery

import (
	"testing"
	"time"
)

func TestNormalizeNextDrawAt(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)

	tests := []struct {
		name string
		base time.Time
		now  time.Time
		want time.Time
	}{
		{
			name: "ignore date and use today slot when still in future",
			base: time.Date(2026, 3, 19, 21, 30, 0, 0, loc),
			now:  time.Date(2026, 3, 18, 20, 0, 0, 0, loc),
			want: time.Date(2026, 3, 18, 21, 30, 0, 0, loc),
		},
		{
			name: "today slot passed then schedule tomorrow",
			base: time.Date(2026, 3, 18, 19, 0, 0, 0, loc),
			now:  time.Date(2026, 3, 18, 20, 0, 0, 0, loc),
			want: time.Date(2026, 3, 19, 19, 0, 0, 0, loc),
		},
		{
			name: "long past date still only keeps time of day",
			base: time.Date(2026, 3, 16, 19, 0, 0, 0, loc),
			now:  time.Date(2026, 3, 18, 20, 0, 0, 0, loc),
			want: time.Date(2026, 3, 19, 19, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		got := normalizeNextDrawAt(tt.base, tt.now)
		if !got.Equal(tt.want) {
			t.Fatalf("%s: got=%s want=%s", tt.name, got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
		}
	}
}
