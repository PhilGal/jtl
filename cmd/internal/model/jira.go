package model

import (
	"strings"

	data "github.com/philgal/jtl/cmd/internal/data"
)

type JiraRequestRow struct {
	Jiraticket string
	Timespent  string
	Comment    string
	Started    string
}

type JiraRequest []JiraRequestRow

//NewJiraRequest creates JiraRequest from CsvRecords
func NewJiraRequest(recs *data.CsvRecords) JiraRequest {
	jr := JiraRequest{}
	for _, row := range *recs {
		//Rows with IDs are pushed, don't them into request
		if row.ID == "" {
			req := JiraRequestRow{
				Jiraticket: row.Ticket,
				Started:    row.StartedTs,
				Comment:    row.Comment,
				Timespent:  row.TimeSpent,
			}
			jr = append(jr, req)
		}
	}
	return jr
}

type JiraResponse struct {

	//{
	//   "self": "https://your-domain.atlassian.net/rest/api/2/issue/10010/worklog/10000",
	//   "author": {
	//     "self": "https://your-domain.atlassian.net/rest/api/2/user?accountId=5b10a2844c20165700ede21g",
	//     "accountId": "5b10a2844c20165700ede21g",
	//     "displayName": "Mia Krystof",
	//     "active": false
	//   },
	//   "updateAuthor": {
	//     "self": "https://your-domain.atlassian.net/rest/api/2/user?accountId=5b10a2844c20165700ede21g",
	//     "accountId": "5b10a2844c20165700ede21g",
	//     "displayName": "Mia Krystof",
	//     "active": false
	//   },
	//   "comment": "I did some work here.",
	//   "updated": "2020-04-09T00:28:56.597+0000",
	//   "visibility": {
	//     "type": "group",
	//     "value": "jira-developers"
	//   },
	//   "started": "2020-04-09T00:28:56.595+0000",
	//   "timeSpent": "3h 20m",
	//   "timeSpentSeconds": 12000,
	//   "id": "100028",
	//   "issueId": "10002"
	// }

	Id        string
	IssueId   string
	Timespent string
	Comment   string
	Started   string
	IsSuccess bool
}

type Credentials struct {
	Username string
	Password string
}

func (creds *Credentials) Trim() *Credentials {
	return &Credentials{strings.TrimSpace(creds.Username), strings.TrimSpace(creds.Password)}
}

func (creds *Credentials) IsValid() bool {
	return creds.Username != "" && creds.Password != ""
}
