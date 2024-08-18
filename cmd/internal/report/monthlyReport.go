package report

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"

	"github.com/philgal/jtl/cmd/duration"
	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/csv"
)

// MonthlyReport displays a short month summary of log items grouped by weeks of a month
type MonthlyReport struct {
	weeklyReports      []*WeeklyReport
	reportsByWeekStart map[string]*WeeklyReport
	totalMinutes       int
	totalTasks         int
	totalTasksPushed   int
}

// NewMonthlyReport generates MonthlyReport by extracting weekly-grouped items from all records in the provided data CSV
func NewMonthlyReport(csvRecords csv.CsvRecords) *MonthlyReport {
	mr := &MonthlyReport{}
	//Create weekly reports.
	//Iterate by CSV rows and append new weekly reports based on weekStart/weekEnd dates deducted from the individual records
	for _, r := range csvRecords {
		startedTs, _ := time.ParseInLocation(config.DefaultDateTimePattern, r.StartedTs, time.Local)
		weekStart, weekEnd := weekBoundaries(startedTs)
		wr := mr.weeklyReportByWeekStart(weekStart)
		wr.weekStart = weekStart
		wr.weekEnd = weekEnd
		wr.totalTasks++
		if r.IsPushed() {
			wr.pushedTasks++
		}
		wr.totalMinutes += duration.DurationToMinutes(r.TimeSpent)
	}
	//Summarize totals from weekly reports
	for _, wr := range mr.weeklyReports {
		mr.totalMinutes += wr.totalMinutes
		mr.totalTasks += wr.totalTasks
		mr.totalTasksPushed += wr.pushedTasks
	}
	return mr
}

// Print displays a MonthlyReport to stdout in a form of formatted a table with a header, rows for weekly summary, and a total monthly summary row.
func (r *MonthlyReport) Print() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Week", "Total tasks (pushed)", "Total time"})
	for _, wr := range r.weeklyReports {
		t.AppendRow([]interface{}{
			fmt.Sprintf("%v - %v", wr.weekStart, wr.weekEnd),
			fmt.Sprintf("%v (%v)", wr.totalTasks, wr.pushedTasks),
			duration.MinutesToDurationString(wr.totalMinutes),
		})
	}
	t.AppendFooter(table.Row{
		"Total for: " + config.GetCurrentDataFileName(),
		fmt.Sprintf("%v (%v)", r.totalTasks, r.totalTasksPushed),
		duration.MinutesToDurationString(r.totalMinutes),
	})
	t.Render()
}

func (r *MonthlyReport) weeklyReportByWeekStart(date string) *WeeklyReport {
	if r.reportsByWeekStart == nil {
		r.reportsByWeekStart = map[string]*WeeklyReport{}
	}
	foundReport, isPresent := r.reportsByWeekStart[date]
	if isPresent {
		return foundReport
	}
	newReport := WeeklyReport{}
	r.reportsByWeekStart[date] = &newReport
	r.weeklyReports = append(r.weeklyReports, &newReport)
	return &newReport
}

func weekBoundaries(t time.Time) (string, string) {
	weekStart := t.AddDate(0, 0, int(time.Monday-t.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 4)
	return weekStart.Format(config.DefaultDatePattern), weekEnd.Format(config.DefaultDatePattern)
}
