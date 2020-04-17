package data

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			name: "Reads all rows from not empty file",
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
			name:         "Reads all rows from an empty file",
			file:         CsvFile{Path: "./csv_testdata/empty.csv"},
			rowsExpected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readFile := &CsvFile{Path: tt.file.Path}
			readFile.ReadAll()
			assert.Equal(t, tt.file.Header, readFile.Header, "Headers are not equal!")
			assert.Equal(t, tt.rowsExpected, len(readFile.Records), "Number of rows are not equal!")
			for i := 0; i < tt.rowsExpected; i++ {
				assert.ElementsMatch(t, readFile.Records[i].AsRow(), tt.file.Records[i].AsRow())
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
		{"Create valid", args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "TICKET-1", "jira"}},
			CsvRecord{
				ID:        "1",
				StartedTs: "14 Apr 2020 11:30",
				Comment:   "Row with ID",
				TimeSpent: "10m",
				Ticket:    "TICKET-1",
				Category:  "jira",
			}, nil},
		{"Create record with invalid jira ticket", args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "ticket1", "jira"}}, CsvRecord{}, fmt.Errorf("Invalid CsvRecord!")},
		{"Create record with invalid timeSpent", args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10bns", "TICKET-1", "jira"}}, CsvRecord{}, fmt.Errorf("Invalid CsvRecord!")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCsvRecord(tt.args.rec)
			t.Logf("%v %v", got, err)
			if err != nil {
				fmt.Printf("Expected error: %q", err)
				assert.Equal(t, tt.wantErr, err)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCsvRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}
