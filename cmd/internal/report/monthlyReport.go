package report

import (
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/data"
)

type MonthlyReport struct {
	weeklyReports      []*WeeklyReport
	reportsByWeekStart map[string]*WeeklyReport
}

func NewMonthlyReport(csvRecords data.CsvRecords) *MonthlyReport {
	mr := &MonthlyReport{}
	for _, r := range csvRecords {
		startedTs, _ := time.ParseInLocation(config.DefaultDateTimePattern, r.StartedTs, time.Local)
		weekStart, weekEnd := weekBoundaries(startedTs)
		wr := mr.weeklyReportByWeekStart(weekStart)
		wr.weekStart = weekStart
		wr.weekEnd = weekEnd
		wr.totalTasks++
		tsm, err := timeSpentToMinutes(r.TimeSpent)
		if err != nil {
			log.Println("Unable to convert timeSpent to minutes!", err)
		}
		wr.totalMinutes += tsm
		// wr.hoursLeft = wr._dueHours - wr.totalHours
	}
	return mr
}

func (r *MonthlyReport) Print() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Week Start Date", "Week End Date", "Total tasks", "Total time"})
	for _, wr := range r.weeklyReports {
		t.AppendRow([]interface{}{
			wr.weekStart,
			wr.weekEnd,
			wr.totalTasks,
			wr.totalTime(),
		})
	}
	t.AppendFooter(table.Row{config.DataFilePath(), "Total", r.totalTasks(), r.totalTime()})
	t.Render()
}

func (r *MonthlyReport) totalTime() string {
	var totalMinutes int
	for _, wr := range r.weeklyReports {
		totalMinutes += wr.totalMinutes
	}
	return minutesToDurationString(totalMinutes)
}

func (r *MonthlyReport) totalTasks() int {
	var totalTasks int
	for _, wr := range r.weeklyReports {
		totalTasks += wr.totalTasks
	}
	return totalTasks
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
