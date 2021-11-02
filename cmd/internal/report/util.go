package report

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/philgal/jtl/cmd/internal/config"
)

func minutesToDurationString(minutes int) string {
	durationString := (time.Duration(minutes) * time.Minute).String()
	return strings.TrimSuffix(strings.TrimSuffix(durationString, "0s"), "0S")
}

func timeSpentToMinutes(timeSpent string) (int, error) {

	//2d, 4h, 2h 30m, 1d 7h 40m
	//TODO add restrictions for 1h = 60m, ...
	sub := strings.SplitN(timeSpent, " ", 2)
	if len(sub) > 1 {
		v0, err := timeSpentToMinutes(sub[0])
		if err != nil {
			return 0, err
		}
		v1, err := timeSpentToMinutes(sub[1])
		return v0 + v1, err
	}
	//1 working day = 8h
	value, _ := strconv.Atoi(strings.TrimRight(timeSpent, "dhm"))
	hour := 60
	day := hour * 8
	if strings.HasSuffix(timeSpent, "d") {
		value = day * value
	} else if strings.HasSuffix(timeSpent, "h") {
		value = hour * value
	} else if strings.HasSuffix(timeSpent, "m") {
	} else {
		return 0, fmt.Errorf("")
	}
	return value, nil
}

func weekBoundaries(t time.Time) (string, string) {
	weekStart := t.AddDate(0, 0, int(time.Monday-t.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 4)
	return weekStart.Format(config.DefaultDatePattern), weekEnd.Format(config.DefaultDatePattern)
}
