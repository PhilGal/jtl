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
	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/data"
	"github.com/philgal/jtl/cmd/internal/report"

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

type aliasReport struct {
	alias      string
	totalTime  string //30h
	totalTasks int
}

var monthlyReport *report.MonthlyReport
var dailyReport *report.DailyReport

func displayReport() {

	csv := data.NewCsvFile(config.DataFilePath())
	csv.ReadAll()
	monthlyReport = report.NewMonthlyReport(csv.Records)
	monthlyReport.Print()

	// csv.Read(data.TodaysRowsCsvRecordPredicate)
	dailyReport = report.NewDailyReport(csv.Records)
	dailyReport.Print()
}
