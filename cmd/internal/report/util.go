package report

import (
	"github.com/philgal/jtl/cmd/internal/config"
	"time"
)

func weekBoundaries(t time.Time) (string, string) {
	weekStart := t.AddDate(0, 0, int(time.Monday-t.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 4)
	return weekStart.Format(config.DefaultDatePattern), weekEnd.Format(config.DefaultDatePattern)
}
