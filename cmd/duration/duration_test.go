package duration

import (
	"reflect"
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
			if got := DateTimeToDate(tt.args.date); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DateTimeToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
