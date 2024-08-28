package duration

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/philgal/jtl/internal/config"
)

type Time struct {
	time.Time
}

const EightHoursInMin = 8 * 60

func ToString(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}
	h := minutes / 60
	m := minutes % 60
	sb := strings.Builder{}
	if h > 0 {
		sb.WriteString(strconv.Itoa(h) + "h")
	}
	if m > 0 {
		if h > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(strconv.Itoa(m) + "m")
	}
	return sb.String()
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
