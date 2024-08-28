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
	"math"
	"os"
	"slices"
	"time"

	"github.com/philgal/jtl/internal/config"
	"github.com/philgal/jtl/internal/csv"
	"github.com/philgal/jtl/internal/duration"
	"github.com/philgal/jtl/internal/model"
	"github.com/spf13/cobra"
)

var (
	// Args
	ticket string
	// Flags
	timeSpent string
	comment   string
	startedTs string
	// Helpers
	now      time.Time = time.Now()
	dayStart           = time.Date(now.Year(), now.Month(), now.Day(), 8, 45, 0, 0, time.Local) // fixme: make day start time configurable
)

const (
	ticketCmdStr  = "ticket"
	timeCmdStr    = "time"
	dateCmdStr    = "date"
	messageCmdStr = "message"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Adds Jira work log occurrence into a data file",
	Long: `
Adds Jira work log occurrence into a data file. Currently, the file is in a CSV format, so it can easily be edited manually before being pushed to a remote Jira <host>.
To save yourself some typing, you can create ticket aliases in a config. Then these aliases can be used instead of ticket ids in the log command with -j flag.
Ticket values specified in a config will the be logged and pushed.

  -----------------------
  %HOME%/.jtl/config.yaml
  -----------------------
  alias:
    jt1: JIRATICKET-1
    l666: ANOTHERLONGTICKET-666
  -----------------------

Examples:
  jtl log -j JIRA-101 -t 30m -s "14 Apr 2020 10:00" -m "Comment"
  jtl log -j l666 -t 1h -s "06 Jun 2020 06:00" -m "Some repeating meeting!"
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ticket = args[0]
		model.ValidateJiraTicketFormat(ticket)
		runLogCommand()
		displayReport()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
	//todo improve duration parsing
	logCmd.Flags().StringVarP(&timeSpent, timeCmdStr, "t", "4h", "[Required] Time spent. Default - 4h")
	logCmd.Flags().StringVarP(&comment, messageCmdStr, "m", "wip", "Comment to the work log. Will be displayed in Jira. Default - \"wip\"")
	logCmd.Flags().StringVarP(&startedTs, dateCmdStr, "d", dayStart.Format(config.DefaultDateTimePattern), "Date and time when the work has been started. Default - current timestamp")
}

var sameDateRecordsFilter = func(logDate time.Time) func(rec csv.Record) bool {
	return func(rec csv.Record) bool {
		return duration.ParseTimeTruncatedToDate(rec.StartedTs).Equal(logDate.Truncate(24 * time.Hour))
	}
}

func runLogCommand() {
	fcsv := csv.NewCsvFile(config.DataFilePath())
	fcsv.ReadAll()

	// total: 4h before pause, 4h after
	// traverse from most recent record
	// sum hours logged in the same day as `startedTs`
	// case =0:  log 4h from 8:45, 4h from 13:00
	// case >0:

	// calculate how much is left of the startedTs
	slices.SortFunc(fcsv.Records, func(a csv.Record, b csv.Record) int {
		return duration.ParseTimeTruncatedToDate(a.StartedTs).Compare(duration.ParseTimeTruncatedToDate(b.StartedTs))
	})

	logDate, _ := time.Parse(config.DefaultDateTimePattern, startedTs)
	//if dates are equal, count hours
	sameDateRecs := fcsv.Filter(sameDateRecordsFilter(logDate))
	minutesSpentToDate := timeSpentToDateInMin(sameDateRecs, logDate)

	// for example 500 > 480 -> 20m to log
	// todo: make max daily duration configurable
	if duration.ToMinutes(timeSpent)+minutesSpentToDate >= duration.EightHoursInMin {
		// todo: make this logic optional if some "distribute" flag is set
		adjustableRecords := csv.Filter(sameDateRecs, func(r csv.Record) bool { return r.ID == "" })
		if len(adjustableRecords) == 0 {
			fmt.Printf("You have already logged %s, will not log more\n", duration.ToString(minutesSpentToDate))
			return
		}

		// calc timeSpent for each record
		totalTimeSpentToLog := math.Min(float64(minutesSpentToDate+duration.ToMinutes(timeSpent)), duration.EightHoursInMin)
		totalRecordsToLog := len(sameDateRecs) + 1
		timeSpentPerRec := int(totalTimeSpentToLog / float64(totalRecordsToLog))
		timeSpent = duration.ToString(timeSpentPerRec)

		for idx, r := range adjustableRecords {
			// calc startedTs for each record
			if idx > 0 {
				prevRec := adjustableRecords[idx-1]
				startedTs, _ := time.Parse(config.DefaultDateTimePattern, prevRec.StartedTs)
				startedTs = startedTs.Add(time.Duration(timeSpentPerRec) * time.Minute)
				r.StartedTs = startedTs.Format(config.DefaultDateTimePattern)
			}
			// set the values
			r.TimeSpent = duration.ToString(timeSpentPerRec)
			fcsv.UpdateRecord(r)
		}
	} else {
		// we have some time to log: do dynamic timeSpent & startedTs calculation for all today's records
		// if today's records are empty, fill 8h with multiple records of 4h.
		// if today's records are not empty, calculate timeSpent and startedTs based on the existing records
		timeSpentMin := int(math.Min(float64(duration.EightHoursInMin-minutesSpentToDate), duration.EightHoursInMin/2))
		timeSpent = duration.ToString(timeSpentMin)
		fmt.Printf("Time spent will is trimmed to %s, to not to exceed %s\n",
			timeSpent,
			duration.ToString(duration.EightHoursInMin))
	}

	// adjust startedTs to the last record: new startedTs = last rec.StartedTs + calculated time spent
	sameDateRecs = fcsv.Filter(sameDateRecordsFilter(logDate))
	if l := len(sameDateRecs); l > 0 {
		lastRec := sameDateRecs[l-1]
		lastRecStaredAt := duration.ParseTime(lastRec.StartedTs)
		startedTs = lastRecStaredAt.Add(time.Minute * time.Duration(duration.ToMinutes(lastRec.TimeSpent))).Format(config.DefaultDateTimePattern)
	}

	if timeSpent != "0m" {
		fcsv.AddRecord(csv.Record{
			ID:        "",
			StartedTs: startedTs,
			Comment:   comment,
			TimeSpent: timeSpent,
			Ticket:    ticket,
		})
		log.Print("Writing a file ", fcsv.Records)
		fcsv.Write()
	} else {
		fmt.Println("Calculated time spent is 0m, will not log!")
		os.Exit(1)
	}
}

func timeSpentToDateInMin(sameDateRecs []csv.Record, logDate time.Time) int {
	var totalTimeSpentOnDate int
	for _, rec := range slices.Backward(sameDateRecs) {
		recDate, _ := time.Parse(config.DefaultDateTimePattern, rec.StartedTs)
		// logs files are already collected by months of the year, so it's enough to compare days
		if recDate.Day() == logDate.Day() {
			totalTimeSpentOnDate += duration.ToMinutes(rec.TimeSpent)
		} else {
			// items are sorted, so first occurrence of mismatched date means there will no more matches.
			return totalTimeSpentOnDate
		}
	}
	return totalTimeSpentOnDate
}

func sameDateRecords(recs []csv.Record, logDate time.Time) []csv.Record {
	return csv.Filter(recs, func(rec csv.Record) bool {
		return duration.ParseTimeTruncatedToDate(rec.StartedTs).Equal(logDate.Truncate(24 * time.Hour))
	})
}
