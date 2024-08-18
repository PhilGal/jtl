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
	"errors"
	"log"
	"slices"
	"time"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/csv"
	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		//use App for testing
		runLogCommand(cmd, args)
		displayReport()
	},
}

const (
	ticketCmdStr  = "ticket"
	timeCmdStr    = "time"
	dateCmdStr    = "date"
	messageCmdStr = "message"
)

func init() {
	rootCmd.AddCommand(logCmd)
	//todo improve duration parsing
	logCmd.Flags().StringP(timeCmdStr, "t", "4h", "[Required] Time spent. Default - 4h")
	logCmd.Flags().StringP(messageCmdStr, "m", "wip", "Comment to the work log. Will be displayed in Jira. Default - \"wip\"")
	logCmd.Flags().StringP(dateCmdStr, "d", time.Now().Format(config.DefaultDateTimePattern), "Date and time when the work has been started. Default - current timestamp")
}

func runLogCommand(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		log.Fatal("too few arguments")
	}

	var ticket, timeSpent, date, comment string

	ticket = args[0]
	if len(ticket) < 5 {
		log.Fatalf("invalid ticket %s", ticket)
	}

	timeSpent, _ = cmd.Flags().GetString(timeCmdStr)
	comment, _ = cmd.Flags().GetString(messageCmdStr)
	date, _ = cmd.Flags().GetString(dateCmdStr)
	fcsv := csv.NewCsvFile(config.DataFilePath())
	fcsv.ReadAll()

	// total: 4h before pause, 4h after
	// traverse from most recent record
	// sum hours logged in the same day as `date`
	// case =0:  log 4h from 8:45, 4h from 13:00
	// case >0:

	// calculate how much is left of the date
	slices.SortFunc(fcsv.Records, func(a csv.CsvRec, b csv.CsvRec) int {
		atime, aerr := time.Parse(config.DefaultDateTimePattern, a.StartedTs)
		btime, berr := time.Parse(config.DefaultDateTimePattern, b.StartedTs)
		if err := errors.Join(aerr, berr); err != nil {
			log.Fatal("cannot compare log records: ", err)
		}
		return atime.Compare(btime)
	})

	fcsv.AddRecord(csv.CsvRec{
		ID:        "",
		StartedTs: date,
		Comment:   comment,
		TimeSpent: timeSpent,
		Ticket:    ticket,
	})

	log.Print(fcsv.Records)
	fcsv.Write()
}
