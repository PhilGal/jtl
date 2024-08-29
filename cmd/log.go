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
	"github.com/philgal/jtl/internal/config"
	"github.com/philgal/jtl/internal/log"
	"github.com/philgal/jtl/internal/model"
	"github.com/spf13/cobra"
)

var (
	// Args
	ticket string
	// Flags
	timeSpent   string
	comment     string
	startedTs   string
	autoFitting bool
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
		executorArgs := log.ExecutorArgs{
			Ticket:    ticket,
			TimeSpent: timeSpent,
			Comment:   comment,
			StartedTs: startedTs}
		if autoFitting {
			log.AutoFitting{ExecutorArgs: executorArgs}.Execute()
		} else {
			log.Normal{ExecutorArgs: executorArgs}.Execute()
		}
		displayReport()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&timeSpent, timeCmdStr, "t", config.DefaultTicketDuration, "[Required] Time spent. Default - 4h")
	logCmd.Flags().StringVarP(&comment, messageCmdStr, "m", "wip", "Comment to the work log. Will be displayed in Jira. Default - \"wip\"")
	logCmd.Flags().StringVarP(&startedTs, dateCmdStr, "d", config.DefaultDayStart, "Date and time when the work has been started. Default - 8:45")
	logCmd.Flags().BoolVarP(&autoFitting, "auto-fitting", "f", true, "Auto-fittimg mode adjusts not pushed records to fit the maximum *daily* duration. If false - logs whatever the input is! Default - true")
}
