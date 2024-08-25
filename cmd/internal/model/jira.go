package model

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/philgal/jtl/cmd/internal/csv"
	"github.com/spf13/viper"
)

type JiraRequestRow struct {
	_rowIdx    int
	Jiraticket string
	Timespent  string
	Comment    string
	Started    string
}

func (rr *JiraRequestRow) GetIdx() int {
	return rr._rowIdx
}

type JiraRequest []JiraRequestRow

// NewJiraRequest creates JiraRequest from CsvRecords
func NewJiraRequest(recs *csv.CsvRecords) JiraRequest {
	jr := JiraRequest{}
	for _, row := range recs.Filter(func(r csv.CsvRec) bool { return !r.IsPushed() }) {
		//Rows with IDs are pushed, don't them into request
		req := JiraRequestRow{
			_rowIdx:    row.GetIdx(),
			Jiraticket: row.Ticket,
			Started:    row.StartedTs,
			Comment:    row.Comment,
			Timespent:  row.TimeSpent,
		}
		jr = append(jr, req)
	}
	return jr
}

type JiraResponse struct {
	RowIdx    int
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

func ValidateJiraTicketFormat(ticket string) {
	pkeyPattern := viper.GetString("projectkeypattern")
	rx := regexp.MustCompile(pkeyPattern)
	if !rx.MatchString(ticket) {
		fmt.Printf("Ticket (project key) %s must match pattern %s\n", ticket, pkeyPattern)
		os.Exit(1)
	}
}
