// Copyright © 2020 Philipp Galichkin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	// _ "github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/table"
	data "github.com/philgal/jtl/cmd/internal/data"

	//TODO impl
	_ "github.com/philgal/jtl/cmd/internal/report"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Displays summarized report for data file",
	Run: func(cmd *cobra.Command, args []string) {
		displayReport()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}

type monthlyReport struct {
	weeklyReports      []*weeklyReport
	reportsByWeekStart map[string]*weeklyReport
}

func (r *monthlyReport) totalTime() string {
	var totalMinutes int
	for _, wr := range r.weeklyReports {
		totalMinutes += wr.totalMinutes
	}
	return minutesToDurationString(totalMinutes)
}

func (r *monthlyReport) totalTasks() int {
	var totalTasks int
	for _, wr := range r.weeklyReports {
		totalTasks += wr.totalTasks
	}
	return totalTasks
}

func (r *monthlyReport) weeklyReportByWeekStart(date string) *weeklyReport {
	if r.reportsByWeekStart == nil {
		r.reportsByWeekStart = map[string]*weeklyReport{}
	}
	foundReport, isPresent := r.reportsByWeekStart[date]
	if isPresent {
		return foundReport
	}
	newReport := weeklyReport{}
	r.reportsByWeekStart[date] = &newReport
	r.weeklyReports = append(r.weeklyReports, &newReport)
	return &newReport
}

type weeklyReport struct {
	weekStart    string
	weekEnd      string
	totalTasks   int //including aliased, len(records)
	totalMinutes int
	_dueHours    int //from configuration, how many ours it is required to log
	hoursLeft    int //number dueHours - totalHours, convert to human readable time
	aliasReports []aliasReport
}

func (wr *weeklyReport) totalTime() string {
	return minutesToDurationString(wr.totalMinutes)
}

func minutesToDurationString(minutes int) string {
	durationString := time.Duration(time.Duration(minutes) * time.Minute).String()
	return strings.TrimSuffix(strings.TrimSuffix(durationString, "0s"), "0S")
}

type aliasReport struct {
	alias      string
	totalTime  string //30h
	totalTasks int
}

var mr *monthlyReport

func displayReport() {
	csv := data.NewCsvFile(dataFile)
	csv.ReadAll()
	mr = buildMonthlyReport(csv.Records)
	printMonthlyReport(mr)
}

func buildMonthlyReport(csvRecords data.CsvRecords) *monthlyReport {
	mr = &monthlyReport{}
	for _, r := range csvRecords {
		startedTs, _ := time.ParseInLocation(defaultDateTimePattern, r.StartedTs, time.Local)
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

func timeSpentToMinutes(timeSpent string) (int, error) {

	//2d, 4h, 2h 30m, 1d 7h 40m
	//TODO add restrictions for 1h = 60m, ...
	sub := strings.SplitN(timeSpent, " ", 2)
	if len(sub) > 1 {
		v0, err := timeSpentToMinutes(sub[0])
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

	var weekStart time.Time

	addDays := func(t time.Time, daysToAdd int) time.Time {
		return t.AddDate(0, 0, daysToAdd)
	}
	calculateWeekEnd := func(t time.Time) time.Time {
		return addDays(t, 4)
	}

	switch t.Weekday() {
	case time.Monday:
		weekStart = t
	case time.Tuesday:
		weekStart = addDays(t, -1)
	case time.Wednesday:
		weekStart = addDays(t, -2)
	case time.Thursday:
		weekStart = addDays(t, -3)
	case time.Friday:
		weekStart = addDays(t, -4)
	}

	return weekStart.Format(defaultDatePattern), calculateWeekEnd(weekStart).Format(defaultDatePattern)
}

func printMonthlyReport(mr *monthlyReport) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Week Start Date", "Week End Date", "Total tasks", "Total time"})
	for _, wr := range mr.weeklyReports {
		t.AppendRow([]interface{}{
			wr.weekStart,
			wr.weekEnd,
			wr.totalTasks,
			wr.totalTime(),
		})
	}
	t.AppendFooter(table.Row{dataFileName(), "Total", mr.totalTasks(), mr.totalTime()})
	t.Render()
}
