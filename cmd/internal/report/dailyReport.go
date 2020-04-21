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

type DailyReport struct {
	totalTasks         int
	timeSpentInMinutes int
	csvRecords         data.CsvRecords
}

func NewDailyReport(csvRecords data.CsvRecords) *DailyReport {
	dr := &DailyReport{}
	dr.csvRecords = csvRecords
	for _, r := range csvRecords {
		if !data.TodaysRowsCsvRecordPredicate(r) {
			continue
		}
		dr.totalTasks++
		tsm, err := timeSpentToMinutes(r.TimeSpent)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		dr.timeSpentInMinutes += tsm
		// }
	}
	return dr
}

func (r *DailyReport) timeSpent() string {
	return minutesToDurationString(r.timeSpentInMinutes)
}

func (r *DailyReport) Print() {
	log.Println(r)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"started at", "ticket", "time tracked", "comment", "pushed to Jira?"})
	var totalPushed int
	for _, rec := range r.csvRecords {
		if !data.TodaysRowsCsvRecordPredicate(rec) {
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

	t.AppendFooter(table.Row{"today: " + time.Now().Format(config.DefaultDatePattern), "", r.timeSpent(), "", fmt.Sprintf("%v/%v", totalPushed, r.totalTasks)})
	t.Render()
}
