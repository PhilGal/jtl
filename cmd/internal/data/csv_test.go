package data

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCsvFile_Write(t *testing.T) {
	t.Run("Write empty file", func(t *testing.T) {
		file := CsvFile{Path: "test.csv"}
		file.Write()
		if !fileExists(file.Path) {
			t.Error("File does not exist")
		}
	})

	t.Run("Test write-read not empty file", func(t *testing.T) {
		newFile := CsvFile{
			Path:    "test.csv",
			Header:  GetCsvHeader(),
			Records: CsvRecords{NewCsvRecord([]string{"1", "14 Apr 2020 11:30", "US demo", "10m", "TICKET-1", "tik"})},
		}
		newFile.Write()

		//read the same and assert fields
		f := NewCsvFile(newFile.Path)
		f.Read()

		assert := assert.New(t)
		assert.Equal("1", f.Records[0].ID)
		assert.Equal("14 Apr 2020 11:30", f.Records[0].StartedTs)
		assert.Equal("US demo", f.Records[0].Comment)
		assert.Equal("10m", f.Records[0].TimeSpent)
		assert.Equal("TICKET-1", f.Records[0].Ticket)
		assert.Equal("tik", f.Records[0].Category)

	})
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
