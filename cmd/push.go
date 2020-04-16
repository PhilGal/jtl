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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	data "github.com/philgal/jtl/cmd/internal/data"
	model "github.com/philgal/jtl/cmd/internal/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes data to a remote Jira server defined as <host> in config.yaml.",
	Long: `Pushes data to a remote server defined as <host> in config.yaml.
For correct work this command reqiures a configured <host> and user <credentials> in the config.yaml:

  host: https://jira.server.url
	credentials:
	  username: <username>
	  password: <password>

However, if username and password are not defined, a user will be prompted to enter them.

Preview mode:
To make sure the data to be pushed is correct, the command can be executed with -p flag. 
The preview output contains host, username and prepared requests bodies for POST request to Jira. 
	`,
	Run: func(cmd *cobra.Command, args []string) {
		PushToServer(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().BoolP("preview", "p", false, "Preview request to be sent to Jira server")
}

const jiraURLTemplate = "/rest/api/2/issue/%v/worklog"

var creds model.Credentials

//PushToServer reads report data and logs work on jira server
func PushToServer(cmd *cobra.Command, args []string) {
	// _, data, _ := data.ReadCsv(dataFile)
	preview := func(jr model.JiraRequest) {
		fmt.Printf("------------\n%v\n------------\n", "PREVIEW MODE")
		fmt.Printf("Jira server: %v\n", viper.GetString("host"))
		fmt.Println("User:", readCredentials().Username)
		for _, row := range jr {
			fmt.Println()
			fmt.Println("POST", buildPostURL(row.Jiraticket))
			fmt.Println(jsonBodyStr(&row))
			fmt.Println()
		}
		fmt.Printf("Total requests: %v\n", len(jr))
		fmt.Printf("-----\n%v\n-----\n", "Done!")
	}
	post := func(cred *model.Credentials, jiraReq model.JiraRequest) []model.JiraResponse {
		client := &http.Client{}
		client.Timeout = time.Second * 30
		responses := []model.JiraResponse{}
		for _, row := range jiraReq {
			req, _ := buildHTTPRequest(row.Jiraticket, cred, &row)
			res, err := client.Do(req)
			if err != nil {
				log.Fatalf("Failed to send %v: %v\n", req, err)
			}
			defer res.Body.Close()
			body, err := httputil.DumpResponse(res, true)
			if err != nil {
				log.Fatal(err)
			}
			//if response was successful
			if res.StatusCode == 201 {
				//unmarshall response
				var jiraRes model.JiraResponse
				err = json.Unmarshal(body, &jiraRes)
				if err != nil {
					log.Println("Error unmarshalling json:", err)
				}
				responses = append(responses, jiraRes)
			} else {
				responses = append(responses, model.JiraResponse{IsSuccess: false})
			}
			//TODO: print logs to file
			fmt.Printf("Jira responded: %v\n{%q}\n", res.Status, body)
		}
		return responses
	}
	csvFile := data.NewCsvFile(dataFile)
	csvFile.Read()
	csvRecords := csvFile.Records
	jreq := model.NewJiraRequest(&csvRecords)
	if shouldPreview, _ := cmd.Flags().GetBool("preview"); shouldPreview == true {
		preview(jreq)
	} else {
		resp := post(readCredentials(), jreq)
		updatePushedRecordsIds(resp, &csvFile)
	}
}

func updatePushedRecordsIds(resp []model.JiraResponse, file *data.CsvFile) {
	csvRecords := file.Records
	if len(resp) != len(file.Records) {
		fmt.Println("[Warning] Mismatch between CSV records and Jira response.")
	}
	for i, csvRec := range csvRecords {
		if resp[i].IsSuccess {
			csvRec.ID = resp[i].Id
		}
	}
	file.Records = csvRecords
	file.Write()
}

func buildHTTPRequest(jiraTicket string, cred *model.Credentials, jr *model.JiraRequestRow) (*http.Request, error) {
	jsonBody := []byte(jsonBodyStr(jr))
	req, err := http.NewRequest("POST", buildPostURL(jiraTicket), bytes.NewBuffer(jsonBody))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", basicAuth(cred)))
	req.Header.Add("Content-Type", "application/json")
	return req, err
}

func buildPostURL(jiraTicket string) string {
	return strings.TrimSuffix(viper.GetString("Host"), "/") + fmt.Sprintf(jiraURLTemplate, jiraTicket)
}

func jsonBodyStr(jr *model.JiraRequestRow) string {
	jsonBodyTemplate := `{"timeSpent": "%v", "comment":"%v", "started": "%v"}`
	return fmt.Sprintf(jsonBodyTemplate, jr.Timespent, jr.Comment, convertDateToDateTimeIso(jr.Started))
}

//convertDateToIDateTimeIso converts date "02-01-2006" to iso "2006-01-02T15:04:05.000-0700"
func convertDateToDateTimeIso(date string) string {
	parsedDate, err := time.ParseInLocation("02 Jan 2006 15:04", date, time.Local)
	if err != nil {
		log.Fatal(err)
	}
	const iso = "2006-01-02T15:04:05.000-0700"
	return parsedDate.Format(iso)
}

func basicAuth(cred *model.Credentials) string {
	auth := cred.Username + ":" + cred.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func readCredentials() *model.Credentials {
	//Read from config first
	err := viper.UnmarshalKey("credentials", &creds)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	creds = *creds.Trim()
	if creds.IsValid() {
		return &creds
	}
	//Otherwise read from user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Username: ")
	creds.Username, _ = reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err == nil {
		fmt.Println("\nPassword typed: " + string(bytePassword))
	}
	creds.Password = string(bytePassword)
	creds = *creds.Trim()
	return &creds
}
