package duration

import (
	"testing"
	"time"
)

func TestDateTimeToDate(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"TestDateTimeToDate",
			args{"14 Apr 2020 10:00"},
			time.Date(2020, 4, 14, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTimeTruncatedToDate(tt.args.date); got.Compare(Time{tt.want}) != 0 {
				t.Errorf("ParseTimeTruncatedToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinutesToDuraion(t *testing.T) {
	tests := []struct {
		name    string
		minutes int
		want    string
	}{
		{"0m", 0, "0m"},
		{"1m", 1, "1m"},
		{"1h", 60, "1h"},
		{"2h 40m", 160, "2h 40m"},
		{"8h", 480, "8h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.minutes); got != tt.want {
				t.Errorf("ToString(%v) = %v, want %v", tt.minutes, got, tt.want)
			}
		})
	}
}
