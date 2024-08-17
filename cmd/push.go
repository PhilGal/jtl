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
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/cmd/internal/csv"
	"github.com/philgal/jtl/cmd/internal/model"
	"github.com/philgal/jtl/cmd/internal/rest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
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
		PushToServer(cmd)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().BoolP("preview", "p", false, "Preview request to be sent to Jira server")
}

const jiraURLTemplate = "/rest/api/2/issue/%v/worklog"

var creds model.Credentials

// PushToServer reads report data and logs work on jira server
func PushToServer(cmd *cobra.Command) {
	push(cmd, rest.HTTPClient)
	displayReport()
}

func push(cmd *cobra.Command, restClient rest.Client) {
	csvFile := csv.NewCsvFile(config.DataFilePath())
	csvFile.ReadAll()
	jreq := model.NewJiraRequest(&csvFile.Records)

	if viper.GetString("Host") == "" {
		fmt.Println("Jira host is not set in config, printing preview")
		preview(jreq)
		return
	}

	if shouldPreview, _ := cmd.Flags().GetBool("preview"); shouldPreview {
		preview(jreq)
		return
	}

	resp := post(readCredentials(), jreq, restClient)
	updatePushedRecordsIds(resp, csvFile.Records)
	log.Printf("CSV records, updated after push: %q\n", csvFile.Records)
	csvFile.Write()
}

func preview(jr model.JiraRequest) {
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

func post(cred *model.Credentials, jiraReq model.JiraRequest, restClient rest.Client) []model.JiraResponse {
	responses := []model.JiraResponse{}
	respChn := make(chan model.JiraResponse, len(jiraReq))
	wg := sync.WaitGroup{}
	for _, row := range jiraReq {
		wg.Add(1)
		go postSingleRequest(cred, row, restClient, respChn, &wg)
	}
	wg.Wait()
	close(respChn)
	for resp := range respChn {
		responses = append(responses, resp)
	}
	return responses
}

func postSingleRequest(cred *model.Credentials, row model.JiraRequestRow, restClient rest.Client, respChn chan model.JiraResponse, wg *sync.WaitGroup) {
	req, _ := buildHTTPRequest(row.Jiraticket, cred, &row)
	res, err := restClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to send %v: %v\n", req, err)
		fmt.Printf("Failed to send %v: %v\n", req, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	//if response was successful
	jiraRes := model.JiraResponse{RowIdx: row.GetIdx(), IsSuccess: false}
	if res.StatusCode == 201 {
		//unmarshall response
		err = json.Unmarshal(body, &jiraRes)
		if err != nil {
			log.Println("Error unmarshalling json:", err)
		}
		jiraRes.IsSuccess = true
	}

	log.Printf("Jira server responded: %v\n{%q}\n", res.Status, body)
	respChn <- jiraRes
	wg.Done()
}

func updatePushedRecordsIds(resp []model.JiraResponse, csvRecords csv.CsvRecords) {
	if len(resp) == 0 {
		return
	}
	log.Println("Generated JiraResponses:\n", resp)
	log.Printf("CSV records before update: %q\n", csvRecords)
	for _, responseItem := range resp {
		if responseItem.IsSuccess {
			csvRecords[responseItem.RowIdx].ID = responseItem.Id
			log.Printf("Updated CSV record %q with ID %v\n", csvRecords[responseItem.RowIdx], responseItem.Id)
		}
	}
}

func buildHTTPRequest(jiraTicket string, cred *model.Credentials, jr *model.JiraRequestRow) (*http.Request, error) {
	jsonBody := []byte(jsonBodyStr(jr))
	req, err := http.NewRequest("POST", buildPostURL(jiraTicket), bytes.NewBuffer(jsonBody))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", basicAuth(cred)))
	req.Header.Add("Content-Type", "application/json")
	log.Println("[Prepared HTTP Request]\n", req)
	return req, err
}

func buildPostURL(jiraTicket string) string {
	return strings.TrimSuffix(viper.GetString("Host"), "/") + fmt.Sprintf(jiraURLTemplate, jiraTicket)
}

func jsonBodyStr(jr *model.JiraRequestRow) string {
	jsonBodyTemplate := `{"timeSpent": "%v", "comment":"%v", "started": "%v"}`
	return fmt.Sprintf(jsonBodyTemplate, jr.Timespent, jr.Comment, convertDateToDateTimeIso(jr.Started))
}

// convertDateToIDateTimeIso converts date "02 Jan 2006 15:04" to iso "2006-01-02T15:04:05.000-0700"
func convertDateToDateTimeIso(date string) string {
	parsedDate, err := time.ParseInLocation(config.DefaultDateTimePattern, date, time.Local)
	if err != nil {
		fmt.Printf("Error! %v", fmt.Errorf("couldn't parse date, %w", err))
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
		log.Fatalf("Unable to decode into struct, %v", err)
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
	bytePassword, err := term.ReadPassword(0)
	if err == nil {
		fmt.Println("\nError reading password from user: ", err)
		log.Println("\nError reading password from user: ", err)
	}
	creds.Password = string(bytePassword)
	creds = *creds.Trim()
	return &creds
}
