package report

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"

	"github.com/philgal/jtl/internal/config"
	"github.com/philgal/jtl/internal/csv"
	"github.com/philgal/jtl/internal/duration"
)

// DailyReport represents a day summary and individual logged tickets
type DailyReport struct {
	showAll                 bool
	tasksToday              int
	totalTasks              int
	timeSpentInMinutesToday int
	timeSpentInMinutes      int
	csvRecords              []csv.Record
}

// NewDailyReport generates DailyReport by extracting today's data from all records in the provided data CSV.
func NewDailyReport(csvRecords []csv.Record, showAll bool) *DailyReport {
	dr := &DailyReport{}
	dr.showAll = showAll
	dr.csvRecords = csvRecords
	for _, r := range csvRecords {
		if csv.TodaysRowsCsvRecordPredicate(r) {
			dr.tasksToday++
			dr.timeSpentInMinutesToday = addTimeSpent(r, dr.timeSpentInMinutesToday)
		}
		dr.totalTasks++
		dr.timeSpentInMinutes = addTimeSpent(r, dr.timeSpentInMinutes)
	}
	return dr
}

// Print displays DailyReport to stdout in a form of formatted a table with a header, rows for individual logs, and a summary row.
// It also displays if the log item has been pushed to the Jira server, and number of pushed records out of all today's logs
func (r *DailyReport) Print() {
	log.Println(r)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"started at", "ticket", "time tracked (today)", "comment", "pushed to Jira? (today)"})
	var totalPushed int
	for _, rec := range r.csvRecords {
		if !r.showAll && !csv.TodaysRowsCsvRecordPredicate(rec) {
			continue
		}
		var isPushed string
		if rec.IsPushed() {
			isPushed = "Y"
			totalPushed++
		} else {
			isPushed = "N"
		}
		t.AppendRow(table.Row{rec.StartedTs, rec.Ticket, rec.TimeSpent, rec.Comment, isPushed})
	}

	t.AppendFooter(table.Row{
		"today: " + time.Now().Format(config.DefaultDatePattern),
		"", //ticket
		fmt.Sprintf("%v (%v)",
			duration.ToString(r.timeSpentInMinutes),
			duration.ToString(r.timeSpentInMinutesToday)), //time tracked
		"", //comment
		fmt.Sprintf("%v/%v", totalPushed, r.tasksToday), //pushed to jira
	})
	t.Render()
}

func addTimeSpent(r csv.Record, timeSpentInMinutes int) int {
	tsm := duration.ToMinutes(r.TimeSpent)
	return timeSpentInMinutes + tsm
}
