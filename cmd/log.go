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
	"log"
	"strings"
	"time"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

type logRecord struct {
	time    string
	message string
	date    string
	ticket  string
}

func (rec logRecord) asArray() []string {
	dataRow := make([]string, 6)
	dataRow[1] = rec.date
	dataRow[2] = rec.message
	dataRow[3] = rec.time
	aliaces := viper.GetStringMapString("alias")
	log.Print(aliaces)
	if len(aliaces) > 0 && aliaces[strings.ToLower(rec.ticket)] != "" {
		dataRow[4] = aliaces[strings.ToLower(rec.ticket)]
		dataRow[5] = rec.ticket
	} else {
		dataRow[4] = rec.ticket
		dataRow[5] = "jira"
	}
	log.Print(dataRow)
	return dataRow
}

const (
	ticketCmdStr  = "ticket"
	timeCmdStr    = "time"
	dateCmdStr    = "date"
	messageCmdStr = "message"
)

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().StringP(ticketCmdStr, "j", "", "[Required] Jira ticket. Ticket aliases can be used. See > jtl help log")
	logCmd.MarkFlagRequired(ticketCmdStr)
	//todo improve duration parsing
	logCmd.Flags().StringP(timeCmdStr, "t", "8h", "[Required] Time spent. Default - 8h")
	logCmd.Flags().StringP(messageCmdStr, "m", "", "Comment to the work log. Will be displayed in Jira. Default - empty")
	logCmd.Flags().StringP(dateCmdStr, "d", time.Now().Format(config.DefaultDateTimePattern), "Date and time when the work has been started. Default - current timestamp")
}

func runLogCommand(cmd *cobra.Command, args []string) {
	rec.ticket, _ = cmd.Flags().GetString(ticketCmdStr)
	rec.time, _ = cmd.Flags().GetString(timeCmdStr)
	rec.message, _ = cmd.Flags().GetString(messageCmdStr)
	rec.date, _ = cmd.Flags().GetString(dateCmdStr)
	newRec, err := data.NewCsvRecord(rec.asArray())
	if err != nil {
		log.Fatalln(err)
	} else {
		csv := data.NewCsvFile(config.DataFilePath())
		csv.ReadAll()
		csv.AddRecord(newRec)
		csv.Write()
	}
}
