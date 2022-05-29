package report

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/philgal/jtl/cmd/internal/config"
)

func minutesToDurationString(minutes int) string {
	durationString := (time.Duration(minutes) * time.Minute).String()
	return strings.TrimSuffix(strings.TrimSuffix(durationString, "0s"), "0S")
}

//durationToMinutes converts string duration d "2D", "4h", "2H 30m", "1d 7h 40m", etc, to minutes.
//if it fails to process a duration, it returns (-1, error)
func durationToMinutes(d string) (int, error) {
	duration := strings.ToLower(d)

	sub := strings.SplitN(duration, " ", 2)
	if len(sub) > 1 {
		v0, err := durationToMinutes(sub[0])
		if err != nil {
			return 0, err
		}
		v1, err := durationToMinutes(sub[1])
		return v0 + v1, err
	}
	//TODO add restrictions for 1h = 60m, ...
	durationValue, _ := strconv.Atoi(strings.TrimRight(duration, "dhm"))
	durationUnit, _ := utf8.DecodeLastRuneInString(duration)
	switch durationUnit {
	case 'd':
		return 8 * 60 * durationValue, nil //consider one working day = 8h
	case 'h':
		return 60 * durationValue, nil
	case 'm':
		return durationValue, nil
	default:
		return -1, fmt.Errorf("Invalid duration unit: %c", durationUnit)
	}
}

func weekBoundaries(t time.Time) (string, string) {
	weekStart := t.AddDate(0, 0, int(time.Monday-t.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 4)
	return weekStart.Format(config.DefaultDatePattern), weekEnd.Format(config.DefaultDatePattern)
}
