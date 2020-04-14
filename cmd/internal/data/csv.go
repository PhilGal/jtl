package data

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

//ReadCsv reads CSV from file.  Returns header, rows and error
func ReadCsv(file string) ([]string, [][]string, error) {
	fcsv, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(fcsv)
	//Read header to skip it. Maybe later add a param like "readHeader: bool"
	header, err := reader.Read()
	records, err := reader.ReadAll()
	return header, records, err
}

//WriteRowCsv appends row into a given CSV file
func WriteRowCsv(file string, row []string) {
	header, data, err := ReadCsv(file)
	fcsv, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	row[0] = strconv.Itoa(len(data) + 1)
	data = append(data, row)
	writer := csv.NewWriter(fcsv)
	writer.Comma = ','
	err = writer.Write(header)
	err = writer.WriteAll(data)
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()
}
