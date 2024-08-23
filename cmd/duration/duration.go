package duration

import (
	"fmt"
	"github.com/philgal/jtl/cmd/internal/config"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Time struct {
	time.Time
}

const EightHoursInMin = 8 * 60

func ToString(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}
	durationString := (time.Duration(minutes) * time.Minute).String()
	return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(durationString, "0s"), "0S"), "0m")
}

// ToMinutes converts string duration d "2D", "4h", "2H 30m", "1d 7h 40m", etc, to minutes.
// if it fails to process a duration, it returns (-1, error)
func ToMinutes(d string) int {
	duration := strings.ToLower(d)

	sub := strings.SplitN(duration, " ", 2)
	if len(sub) > 1 {
		v0 := ToMinutes(sub[0])
		v1 := ToMinutes(sub[1])
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

// ParseTime converts a string date into time.Time using custom datetime pattern from the config
func ParseTime(date string) Time {
	t, _ := time.Parse(config.DefaultDateTimePattern, date)
	return Time{t}
}

// ParseTimeTruncatedToDate converts a string date into time.Time using custom datetime pattern from the config, then trims the time part to zeroes
func ParseTimeTruncatedToDate(date string) Time {
	return ParseTime(date).TruncateToDate()
}

func (t Time) TruncateToDate() Time {
	return Time{t.Truncate(24 * time.Hour)}
}

// Overrides

func (t Time) Compare(other Time) int {
	return t.Time.Compare(other.Time)
}
