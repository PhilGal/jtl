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
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		PushToServer(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().BoolP("preview", "p", false, "Preview request to be sent to Jira server")
}

type jiraRequestRow struct {
	rownum     int
	jiraticket string
	timespent  string
	comment    string
	started    string
}

type jiraRequest []jiraRequestRow

type credentials struct {
	Username string
	Password string
}

func (creds *credentials) trim() *credentials {
	return &credentials{strings.TrimSpace(creds.Username), strings.TrimSpace(creds.Password)}
}

func (creds *credentials) isValid() bool {
	return creds.Username != "" && creds.Password != ""
}

const jiraURLTemplate = "/rest/api/2/issue/%v/worklog"

var Creds credentials

//PushToServer reads report data and logs work on jira server
func PushToServer(cmd *cobra.Command, args []string) {
	_, data, _ := readCsv(dataFile)
	jreq := convertCsvDataIntoJiraRequest(data)
	preview := func(jr jiraRequest) {
		fmt.Printf("------------\n%v\n------------\n", "PREVIEW MODE")
		fmt.Printf("Jira server: %v\n", viper.GetString("host"))
		fmt.Println("User:", readCredentials().Username)
		for _, row := range jr {
			fmt.Println()
			fmt.Println("POST", postURL(row.jiraticket))
			fmt.Println(jsonBodyStr(&row))
			fmt.Println()
		}
		fmt.Printf("Total requests: %v\n", len(jreq))
		fmt.Printf("-----\n%v\n-----\n", "Done!")
	}
	post := func(cred *credentials, jr jiraRequest) {
		client := &http.Client{}
		client.Timeout = time.Second * 30
		for _, row := range jr {
			req, _ := buildHTTPRequest(row.jiraticket, cred, &row)
			res, err := client.Do(req)
			if err != nil {
				log.Fatalf("Failed to send %v: %v\n", req, err)
			}
			defer res.Body.Close()
			body, err := httputil.DumpResponse(res, true)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Jira responded: %v {%q}\n", res.Status, body)
		}
	}
	if shouldPreview, _ := cmd.Flags().GetBool("preview"); shouldPreview == true {
		preview(jreq)
	} else {
		post(readCredentials(), jreq)
	}
}

func convertCsvDataIntoJiraRequest(data [][]string) jiraRequest {
	var jr jiraRequest
	for i, row := range data {
		req := &jiraRequestRow{
			i,      //row number
			row[4], //jira ticket
			row[3], //hours -> timespent
			row[2], //comment -> activity
			row[1], //date -> started
		}
		jr = append(jr, *req)
	}
	return jr
}

func buildHTTPRequest(jiraTicket string, cred *credentials, jr *jiraRequestRow) (*http.Request, error) {
	jsonBody := []byte(jsonBodyStr(jr))
	req, err := http.NewRequest("POST", postURL(jiraTicket), bytes.NewBuffer(jsonBody))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", basicAuth(cred)))
	req.Header.Add("Content-Type", "application/json")
	return req, err
}

func postURL(jiraTicket string) string {
	return strings.TrimSuffix(viper.GetString("Host"), "/") + fmt.Sprintf(jiraURLTemplate, jiraTicket)
}

func jsonBodyStr(jr *jiraRequestRow) string {
	jsonBodyTemplate := `{"timeSpent": "%v", "comment":"%v", "started": "%v"}`
	return fmt.Sprintf(jsonBodyTemplate, jr.timespent, jr.comment, convertDateToDateTimeIso(jr.started))
}

//convertDateToIDateTimeIso converts date "02-01-2006" to iso "2006-01-02T15:04:05.000-0700"
func convertDateToDateTimeIso(date string) string {
	parsedDate, err := time.Parse("02 Jan 2006 15:04", date)
	if err != nil {
		log.Fatal(err)
	}
	const iso = "2006-01-02T15:04:05.000-0700"
	return parsedDate.Local().Add(time.Hour * 10).Format(iso)
}

func basicAuth(cred *credentials) string {
	auth := cred.Username + ":" + cred.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func readCredentials() *credentials {
	//Read from config first
	err := viper.UnmarshalKey("credentials", &Creds)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	Creds = *Creds.trim()
	if Creds.isValid() {
		return &Creds
	}
	//Otherwise read from user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Username: ")
	Creds.Username, _ = reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err == nil {
		fmt.Println("\nPassword typed: " + string(bytePassword))
	}
	Creds.Password = string(bytePassword)
	Creds = *Creds.trim()
	return &Creds
}
