package duration

import (
	"fmt"
	"github.com/philgal/jtl/cmd/internal/config"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const EIGHT_HOURS_IN_MIN = 8 * 60

func MinutesToDurationString(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}
	durationString := (time.Duration(minutes) * time.Minute).String()
	return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(durationString, "0s"), "0S"), "0m")
}

// durationToMinutes converts string duration d "2D", "4h", "2H 30m", "1d 7h 40m", etc, to minutes.
// if it fails to process a duration, it returns (-1, error)
func DurationToMinutes(d string) int {
	duration := strings.ToLower(d)

	sub := strings.SplitN(duration, " ", 2)
	if len(sub) > 1 {
		v0 := DurationToMinutes(sub[0])
		v1 := DurationToMinutes(sub[1])
		return v0 + v1
	}
	//TODO add restrictions for 1h = 60m, ...
	durationValue, _ := strconv.Atoi(strings.TrimRight(duration, "dhm"))
	durationUnit, _ := utf8.DecodeLastRuneInString(duration)
	switch durationUnit {
	case 'd':
		return 8 * 60 * durationValue //consider one working day = 8h
	case 'h':
		return 60 * durationValue
	case 'm':
		return durationValue
	default:
		panic(fmt.Sprintf("invalid duration unit: %c", durationUnit))
	}
}

func DateTimeToTime(date string) time.Time {
	t, _ := time.Parse(config.DefaultDateTimePattern, date)
	return t
}

func DateTimeToDate(date string) time.Time {
	d := DateTimeToTime(date).Truncate(24 * time.Hour)
	return d
}
