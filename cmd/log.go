/*
Copyright © 2020 Philipp Galichkin <phil.gal@outlook.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

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
  jtl log -j l666 -t 1h -s "06 Jun 2020 06:00" -m "Some repeatedly meeting!"
`,
	Run: func(cmd *cobra.Command, args []string) {
		//use App for testing
		runLogCommand(cmd, args)
	},
}

type logRecord struct {
	timeSpent   string
	message     string
	dateStarted string
	jiraTicket  string
}

func init() {
	rootCmd.AddCommand(logCmd)
	//`{"timeSpent": "%vh", "comment":"%v", "started": "%v"}`

	//todo improve duration parsing
	logCmd.Flags().StringP("jiraTicket", "j", "", "[Required] Jira ticket. Ticket aliases can be used. See > jtl help log")
	logCmd.MarkFlagRequired("jiraTicket")
	logCmd.Flags().StringP("timeSpent", "t", "1h", "[Required] Time spent. Default - 1h")
	logCmd.MarkFlagRequired("timeSpent")
	logCmd.Flags().StringP("message", "m", "", "Comment to the work log. Will be displayed in Jira. Default - empty")
	logCmd.Flags().StringP("started", "s", time.Now().String(), "Date and time when the work has been started. Default - current timestamp")
}

func runLogCommand(cmd *cobra.Command, args []string) {

	rec.jiraTicket, _ = cmd.Flags().GetString("jiraTicket")
	rec.timeSpent, _ = cmd.Flags().GetString("timeSpent")
	rec.message, _ = cmd.Flags().GetString("message")
	rec.dateStarted, _ = cmd.Flags().GetString("started")
	aliaces := viper.GetStringMapString("alias")
	fmt.Println(aliaces)
	dataRow := make([]string, 6)
	dataRow[1] = rec.dateStarted
	dataRow[2] = rec.message
	dataRow[3] = rec.timeSpent
	if len(aliaces) > 0 {
		alias := rec.jiraTicket
		aliasTicket := aliaces[alias]
		if aliasTicket != "" {
			validateJiraTicketName(aliasTicket)
			dataRow[4] = aliasTicket
			dataRow[5] = alias
		}
	} else {
		validateJiraTicketName(rec.jiraTicket)
		dataRow[4] = rec.jiraTicket
		dataRow[5] = "jira"
	}
	writeRowCsv(dataFile, dataRow)
}

const (
	jiraTicketRegexp = `(([A-Za-z]{1,10})-?)[A-Z]+-\d+`
)

func validateJiraTicketName(val string) {
	match, err := regexp.MatchString(jiraTicketRegexp, val)
	if err != nil {
		log.Fatalf("Error validating Jira ticket name: %v\n", err)
	}
	if !match {
		log.Fatalf("String '%v' does not seem like a valid Jira ticket.\n", val)
	}
}

func readCsv(file string) ([]string, [][]string, error) {
	fcsv, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(fcsv)
	//Read header to skip it. Maybe later add a param like "readHeader: bool"
	header, err := reader.Read()
	records, err := reader.ReadAll()
	return header, records, err
}

func writeRowCsv(file string, row []string) {
	header, data, err := readCsv(file)
	fcsv, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	row[0] = strconv.Itoa(len(data) + 1)
	data = append(data, row)
	writer := csv.NewWriter(fcsv)
	writer.Comma = ','
	err = writer.Write(header)
	err = writer.WriteAll(data)
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()
}