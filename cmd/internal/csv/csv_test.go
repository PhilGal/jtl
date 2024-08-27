package csv

import (
	"github.com/philgal/jtl/cmd/internal/config"
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

var okCsvRecord = Record{
	ID:        "1",
	StartedTs: "14 Apr 2020 11:30",
	Comment:   "Row with ID",
	TimeSpent: "10m",
	Ticket:    "TICKET-1",
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

// Return value, ignore error
func newCsvRecord(idx int, rec []string) Record {
	newRec := newCsvRec(rec)
	newRec._idx = idx
	return newRec
}

//____TESTS____//

func TestCsvFile_Write(t *testing.T) {
	path := "./csv_testdata/write_file_test.csv"
	t.Cleanup(func() { deleteFileIfExists(path) })
	t.Run("Write one row into a file", func(t *testing.T) {
		writtenFile := File{
			Path:   path,
			Header: config.Header(),
			//Records: []Record{newCsvRecord(0, []string{"1", "14 Apr 2020 11:30", "US demo", "10m", "TICKET-1"})},
			Records: []Record{{ID: "1", StartedTs: "14 Apr 2020 11:30", Comment: "US demo", TimeSpent: "10m", Ticket: "TICKET-1"}},
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
		file         File
		rowsExpected int
	}{
		{
			name: "Should reads all rows from not empty file",
			file: File{
				Path:   "./csv_testdata/not_empty.csv",
				Header: config.Header(),
				Records: []Record{
					{_idx: 0, ID: "1", StartedTs: "14 Apr 2020 11:30", Comment: "Row with ID", TimeSpent: "10m", Ticket: "TICKET-1"},
					{_idx: 1, ID: "", StartedTs: "15 Apr 2020 11:30", Comment: "Row without ID", TimeSpent: "10m", Ticket: "TICKET-2"},
				},
			},
			rowsExpected: 2,
		},

		{
			name:         "Should reads all rows from an empty file",
			file:         File{Path: "./csv_testdata/empty.csv"},
			rowsExpected: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readFile := File{Path: test.file.Path}
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
		name   string
		args   args
		want   Record
		panics bool
	}{
		{
			"Should create valid CsvRecord",
			args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "TICKET-1"}},
			okCsvRecord,
			false},
		// TODO: add proper tests for adding a record to a file
		// {
		// 	"Should not create record with invalid jira ticket",
		// 	args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "ticket1"}},
		// 	Record{},
		// 	true},
		// {
		// 	"Should not create record with invalid timeSpent",
		// 	args{[]string{"1", "14 Apr 2020 11:30", "Row with ID", "10bns", "TICKET-1"}},
		// 	Record{},
		// 	true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.panics {
				assert.Panics(t, func() { newCsvRec(test.args.rec) })
			} else {
				got := newCsvRec(test.args.rec)
				t.Logf("%v", got)
				assert.True(t, reflect.DeepEqual(got, test.want))
			}
		})
	}
}
