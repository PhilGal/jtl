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
func newCsvRecord(idx int, rec []string) CsvRecord {
	newRec, _ := NewCsvRecord(rec)
	newRec._idx = idx
	return newRec
}

//____TESTS____//

func TestCsvFile_Write(t *testing.T) {
	path := "./csv_testdata/write_file_test.csv"
	t.Cleanup(func() { deleteFileIfExists(path) })
	t.Run("Write one row into a file", func(t *testing.T) {
		writtenFile := CsvFile{
			Path:    path,
			Header:  GetCsvHeader(),
			Records: CsvRecords{newCsvRecord(0, []string{"1", "14 Apr 2020 11:30", "US demo", "10m", "TICKET-1", "jira"})},
		}
		writtenFile.Write()

		//read the same and assert fields
		readFile := NewCsvFile(writtenFile.Path)
		readFile.ReadAll()

		assert := assert.New(t)
		assert.ElementsMatch(writtenFile.Records[0].AsRow(), readFile.Records[0].AsRow())
	})
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
					newCsvRecord(0, []string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "TICKET-1", "jira"}),
					newCsvRecord(1, []string{"", "15 Apr 2020 11:30", "Row without ID", "10m", "TICKET-2", "jira"}),
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
			readFile := CsvFile{Path: test.file.Path}
			readFile.ReadAll()
			assert.Exactly(t, test.file, readFile)
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

	updatedRec := CsvRecord{
		_idx:      1,
		ID:        "666",
		StartedTs: "14 Apr 2020 11:30",
		Comment:   "Updated Row with ID",
		TimeSpent: "10m",
		Ticket:    "TICKET-666",
		Category:  "jira",
	}

	file := CsvFile{Path: "./csv_testdata/not_empty.csv"}
	file.ReadAll()

	idxOutOfBounds := len(file.Records) + 1

	tests := []struct {
		name          string
		idx           int
		rec           CsvRecord
		expectedError error
	}{
		{"Should update record at the given index", updatedRec._idx, updatedRec, nil},
		{"Should return error if the given index is out of range", idxOutOfBounds, okCsvRecord, errors.New("Index is out of range")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := file.UpdateRecord(test.idx, test.rec)
			assert.EqualValues(t, err, test.expectedError)
			if err == nil {
				assert.Exactly(t, test.rec, file.Records[test.idx])
				assert.Equal(t, test.idx, file.Records[test.idx]._idx)
			}
		})
	}
}
