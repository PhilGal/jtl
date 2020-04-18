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
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	data "github.com/philgal/jtl/cmd/internal/data"
	"github.com/philgal/jtl/cmd/internal/model"
	"github.com/philgal/jtl/cmd/internal/rest"
	"github.com/stretchr/testify/assert"
)

/*
JiraResponses:
 [{1417210 878641 2h  2020-04-17T08:20:00.000+0200 true} {1417211 927955 1h  2020-04-17T10:00:00.000+0200 true} {1417212 942367 3h  2020-04-17T11:00:00.000+0200 true}]
CSV records before update: [{"" "17 Apr 2020 08:20" "" "2h" "EEBATCH-680" "jira"} {"" "17 Apr 2020 10:00" "" "1h" "PSPTK-2" "capex"} {"" "17 Apr 2020 11:00" "" "3h" "EEGLOBAL-16607" "jira"}]
Updated CSV record {"1417210" "17 Apr 2020 08:20" "" "2h" "EEBATCH-680" "jira"} with ID 1417210
Updated CSV record {"1417211" "17 Apr 2020 10:00" "" "1h" "PSPTK-2" "capex"} with ID 1417211
Updated CSV record {"1417212" "17 Apr 2020 11:00" "" "3h" "EEGLOBAL-16607" "jira"} with ID 1417212
Final CSV records: [{"1417210" "17 Apr 2020 08:20" "" "2h" "EEBATCH-680" "jira"} {"1417211" "17 Apr 2020 10:00" "" "1h" "PSPTK-2" "capex"} {"1417212" "17 Apr 2020 11:00" "" "3h" "EEGLOBAL-16607" "jira"}]
*/

var restClient rest.Client

type MockRestClient struct{}

func (c *MockRestClient) Do(req *http.Request) (*http.Response, error) {
	jsonb, _ := ioutil.ReadFile("./cmd_testdata/jira_response.json")
	return &http.Response{
		StatusCode: 201,
		Body:       ioutil.NopCloser(bytes.NewReader(jsonb)),
	}, nil
}

func init() {
	restClient = &MockRestClient{}
}

func TestPost(t *testing.T) {

	// t.Run("Should post correct values", func(t *testing.T) {
	// 	post()
	// })
	csvFile := data.NewCsvFile("./cmd_testdata/not_empty.csv")
	csvFile.ReadAll()

	t.Run("Should unmarshall correct response", func(t *testing.T) {
		jreq := model.NewJiraRequest(&csvFile.Records)
		jres := post(readCredentials(), jreq, restClient)

		expected := []model.JiraResponse{{Id: "666", IssueId: "111", Timespent: "3h", Started: "2020-04-17T11:00:00.000+0200", IsSuccess: true}}

		assert.Equal(t, len(jres), 1, "Bad response size")
		assert.Exactly(t, expected, jres)
	})
}
