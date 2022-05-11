package report

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/data"
)

//DailyReport represents a daily report
type DailyReport struct {
	showAll                 bool
	tasksToday              int
	totalTasks              int
	timeSpentInMinutesToday int
	timeSpentInMinutes      int
	csvRecords              data.CsvRecords
}

//NewDailyReport creates new report from the given CsvRecords
func NewDailyReport(csvRecords data.CsvRecords, showAll bool) *DailyReport {
	dr := &DailyReport{}
	dr.showAll = showAll
	dr.csvRecords = csvRecords
	for _, r := range csvRecords {
		if data.TodaysRowsCsvRecordPredicate(r) {
			dr.tasksToday++
			dr.timeSpentInMinutesToday = addTimeSpent(r, dr.timeSpentInMinutesToday)
		}
		dr.totalTasks++
		dr.timeSpentInMinutes = addTimeSpent(r, dr.timeSpentInMinutes)
	}
	return dr
}

func addTimeSpent(r data.CsvRecord, timeSpentInMinutes int) int {
	tsm, err := timeSpentToMinutes(r.TimeSpent)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return timeSpentInMinutes + tsm
}

func (r *DailyReport) timeSpent(minutes int) string {
	return minutesToDurationString(minutes)
}

//Print implements Printable
func (r *DailyReport) Print() {
	log.Println(r)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"started at", "ticket", "time tracked (today)", "comment", "pushed to Jira? (today)"})
	var totalPushed int
	for _, rec := range r.csvRecords {
		if !r.showAll && !data.TodaysRowsCsvRecordPredicate(rec) {
			continue
		}
		//TODO: extract
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
		fmt.Sprintf("%v (%v)", r.timeSpent(r.timeSpentInMinutes), r.timeSpent(r.timeSpentInMinutesToday)), //time tracked
		"", //comment
		fmt.Sprintf("%v/%v", totalPushed, r.totalTasks), //pushed to jira
	})
	t.Render()
}
