package data

import (
	"log"
	"os"
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
			Records: CsvRecords{NewCsvRecord([]string{"1", "14 Apr 2020 11:30", "US demo", "10m", "TICKET-1", "jira"})},
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
					NewCsvRecord([]string{"1", "14 Apr 2020 11:30", "Row with ID", "10m", "TICKET-1", "jira"}),
					NewCsvRecord([]string{"", "15 Apr 2020 11:30", "Row without ID", "10m", "TICKET-2", "jira"}),
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
