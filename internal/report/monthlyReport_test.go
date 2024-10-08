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

package report

import (
	"testing"

	"github.com/philgal/jtl/internal/csv"
	"github.com/philgal/jtl/internal/duration"
	"github.com/stretchr/testify/assert"
)

func Test_timeSpentToMinutes(t *testing.T) {

	tests := []struct {
		name      string
		timeSpent string
		want      int
	}{
		{"Should parse 1h", "1h", 60},
		{"Should parse 1d 2h", "1d 2h", (8 + 2) * 60},
		{"Should parse 2d 3h 25m", "2d 3h 25m", (16+3)*60 + 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := duration.ToMinutes(tt.timeSpent)
			if got != tt.want {
				t.Errorf("timeSpentToMinutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewMonthlyReport(t *testing.T) {

	csvRecords := []csv.Record{
		//week 13-17 Apr
		{TimeSpent: "3h", Ticket: "JIRA-1", StartedTs: "14 Apr 2020 12:00"},        //"middle" of the week
		{TimeSpent: "1d 1h 35m", Ticket: "JIRA-1", StartedTs: "17 Apr 2020 12:00"}, //last day of the week
		//week 20-24 Apr
		{TimeSpent: "20m", Ticket: "JIRA-2", StartedTs: "20 Apr 2020 12:00"}, // first day of the week
	}

	tests := []struct {
		name       string
		csvRecords []csv.Record
		want       MonthlyReport
	}{
		// TODO: Add test cases
		{"Builds report",
			csvRecords,
			MonthlyReport{
				weeklyReports: []*WeeklyReport{
					{
						weekStart: "13 Apr 2020",
						weekEnd:   "17 Apr 2020",
						//12h35m = 3*60 + 8*60 + 1*60 + 35 = 755
						totalMinutes: 3*60 + 8*60 + 1*60 + 35,
						totalTasks:   2,
					},
					{
						weekStart: "20 Apr 2020",
						weekEnd:   "24 Apr 2020",
						//20m
						totalMinutes: 20,
						totalTasks:   1,
					},
				},
				totalMinutes:     775, //755 + 20
				totalTasks:       3,
				totalTasksPushed: 0,
			},
		},
	}
	assert := assert.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMonthlyReport(tt.csvRecords)
			assert.Equal(tt.want.totalMinutes, got.totalMinutes)
			assert.Equal(tt.want.totalTasks, got.totalTasks)
			assert.Equal(tt.want.totalTasksPushed, got.totalTasksPushed)
			assert.Equal(2, len(got.weeklyReports))
			for idx, wr := range tt.want.weeklyReports {
				assert.Exactly(wr, got.weeklyReports[idx])
			}
		})
	}
}
