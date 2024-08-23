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
