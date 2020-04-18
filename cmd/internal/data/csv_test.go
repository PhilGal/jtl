package data

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/philgal/jtl/validation"
	"github.com/stretchr/testify/assert"
)

func init() {
	validation.InitValidator()
}

var okCsvRecord = CsvRecord{
	ID:        "1",
	StartedTs: "14 Apr 2020 11:30",
	Comment:   "Row with ID",
	TimeSpent: "10m",
	Ticket:    "TICKET-1",
	Category:  "jira",
}

func TestCsvFile_Write(t *testing.T) {
	path := "./csv_testdata/write_file_test.csv"
	t.Cleanup(func() { deleteFileIfExists(path) })
	t.Run("Write one row into a file", func(t *testing.T) {
		writtenFile := CsvFile{
			Path:    path,
			Header:  GetCsvHeader(),
			Records: CsvRecords{newCsvRecord([]string{"1", "14 Apr 2020 11:30", "US demo", "10m", "TICKET-1", "jira"})},
		}
		writtenFile.Write()

		//read the same and assert fields
		readFile := NewCsvFile(writtenFile.Path)
		readFile.ReadAll()

		assert := assert.New(t)
		assert.ElementsMatch(writtenFile.Records[0].AsRow(), readFile.Records[0].AsRow())
	})
}

func deleteFileIfExists(filename string) {
	if fileExists(filename) {
		if err := os.Remove(filename); err != nil {
			log.Fatalln("Cannot remove file!", err)
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//Return value, ignore error
func newCsvRecord(rec []string) CsvRecord {
	newRec, _ := NewCsvRecord(rec)
	return newRec
}

func TestCsvFile_ReadAll(t *testing.T) {
	tests := []struct {
		name         string
		file         CsvFile
		rowsExpected int
	}{
		{
			name: "Should reads all rows from not empty file",
			file: CsvFile{
				Path:   "./csv_testdata/not_empty.csv",
				Header: GetCsvHeader(),
				Records: CsvRecords{
					newCsvRecord([]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "TICKET-1", "jira"}),
					newCsvRecord([]string{"", "15 Apr 2020 11:30", "Row without ID", "10m", "TICKET-2", "jira"}),
				},
			},
			rowsExpected: 2,
		},

		{
			name:         "Should reads all rows from an empty file",
			file:         CsvFile{Path: "./csv_testdata/empty.csv"},
			rowsExpected: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readFile := &CsvFile{Path: test.file.Path}
			readFile.ReadAll()
			assert.Equal(t, test.file.Header, readFile.Header, "Headers are not equal!")
			assert.Equal(t, test.rowsExpected, len(readFile.Records), "Number of rows are not equal!")
			for i := 0; i < test.rowsExpected; i++ {
				assert.ElementsMatch(t, readFile.Records[i].AsRow(), test.file.Records[i].AsRow())
			}

		})
	}
}

func TestNewCsvRecord(t *testing.T) {
	type args struct {
		rec []string
	}
	tests := []struct {
		name    string
		args    args
		want    CsvRecord
		wantErr error
	}{
		{
			"Should create valid CsvRecord",
			args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "TICKET-1", "jira"}},
			okCsvRecord,
			nil},
		{
			"Should not create record with invalid jira ticket",
			args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "ticket1", "jira"}},
			CsvRecord{},
			fmt.Errorf("Invalid CsvRecord! \"Key: 'CsvRecord.Ticket' Error:Field validation for 'Ticket' failed on the 'jiraticket' tag\"")},
		{
			"Should not create record with invalid timeSpent",
			args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10bns", "TICKET-1", "jira"}},
			CsvRecord{},
			fmt.Errorf("Invalid CsvRecord! \"Key: 'CsvRecord.TimeSpent' Error:Field validation for 'TimeSpent' failed on the 'timespent' tag\""),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewCsvRecord(test.args.rec)
			t.Logf("%v %q", got, err)
			if err != nil {
				fmt.Printf("Expected error: %q", err)
				assert.Equal(t, test.wantErr, err)
			} else if !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewCsvRecord() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestUpdateRecord(t *testing.T) {
	tests := []struct {
		name          string
		idx           int
		rec           CsvRecord
		expectedError error
	}{
		{"Should update record at the given index", 1, okCsvRecord, nil},
		{"Should return error if the given index is out of range", 5, okCsvRecord, errors.New("Index is out of range")},
	}

	file := CsvFile{Path: "./csv_testdata/not_empty.csv"}
	file.ReadAll()
	//pre-condition
	if len(file.Records) != 2 {
		t.Error("Number of records in a file must be 2")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := file.UpdateRecord(test.idx, test.rec)
			assert.EqualValues(t, err, test.expectedError)
			if err == nil {
				assert.Exactly(t, file.Records[test.idx], test.rec)
			}
		})
	}
}
