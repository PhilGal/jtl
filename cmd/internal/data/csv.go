package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator"
	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/validation"
)

//CsvHeader represents header in a CSV file
type CsvHeader []string

var header = CsvHeader{"ID", "StartedTs", "Comment", "TimeSpent", "Ticket", "Category"}

//GetCsvHeader returns a CSV-header row
func GetCsvHeader() CsvHeader {
	return header
}

//CsvRecord represents a single record in a CSV file
type CsvRecord struct {
	_idx      int
	ID        string
	StartedTs string
	Comment   string
	// TimeSpent string `validate:"required"`
	// Ticket    string `validate:"required"`
	TimeSpent string `validate:"required,timespent"`
	Ticket    string `validate:"required,jiraticket"`
	Category  string
}

//GetIdx returns a row's index in CSV file
func (r *CsvRecord) GetIdx() int {
	return r._idx
}

//AsRow represents a CSV-writable row
func (r *CsvRecord) AsRow() []string {
	return []string{r.ID, r.StartedTs, r.Comment, r.TimeSpent, r.Ticket, r.Category}
}

//IsPushed returns true is the item has been already pushed to Jira server
func (r *CsvRecord) IsPushed() bool {
	return r.ID != ""
}

//NewCsvRecord created a new CsvRecord from a slice representing a single CSV row
func NewCsvRecord(rec []string) (CsvRecord, error) {
	if len(rec) != len(header) {
		log.Fatalf("Cannot create CsvRecord. Slice size is %v, expected: %v", len(rec), len(header))
	}
	csvRec := CsvRecord{
		ID:        rec[0],
		StartedTs: rec[1],
		Comment:   rec[2],
		TimeSpent: rec[3],
		Ticket:    rec[4],
		Category:  rec[5],
	}

	if err := validation.Validate.Struct(csvRec); err != nil {
		validationError := err.(validator.ValidationErrors)
		return CsvRecord{}, fmt.Errorf("Invalid CsvRecord! %q", validationError)
	}

	return csvRec, nil
}

//CsvRecords is a wrapper on []CsvRecord
type CsvRecords []CsvRecord

//Filter returns a copy of records, filtered using the given recordFilter
func (recs *CsvRecords) Filter(recordFilter CsvRecordPredicate) CsvRecords {
	filteredRecs := CsvRecords{}
	for _, r := range *recs {
		if recordFilter(r) {
			filteredRecs = append(filteredRecs, r)
		}
	}
	return filteredRecs
}

//AsRows converts CsvRecords into 2-d slice representing CSV {records X columns}
func (recs *CsvRecords) AsRows() [][]string {
	rows := [][]string{}
	for _, r := range *recs {
		rows = append(rows, r.AsRow())
	}
	return rows
}

// func (r *CsvRecord) isEqual(other CsvRecord) bool {
// 	//TODO impl
// 	return false
// }

//CsvFile represents a CSV file with header and records. Use Read() and Write() to read and write data from/to disk
type CsvFile struct {
	Path    string
	Header  CsvHeader
	Records CsvRecords
}

//CsvRecordPredicate used as a condition for rows filtering when reading CSV file
type CsvRecordPredicate func(CsvRecord) bool

//AllRowsCsvRecordPredicate is an always true condition. All rows will be processed
var AllRowsCsvRecordPredicate = func(r CsvRecord) bool {
	return true
}

//RowsWithoutIDsCsvRecordPredicate filters rows with IDs. Only rows without IDs will be processed
var RowsWithoutIDsCsvRecordPredicate = func(r CsvRecord) bool {
	return r.ID == ""
}

//TodaysRowsCsvRecordPredicate filters rows with startedTs = today
var TodaysRowsCsvRecordPredicate = func(r CsvRecord) bool {
	startedTs, err := time.Parse(config.DefaultDateTimePattern, r.StartedTs)
	if err != nil {
		fmt.Printf("Error! %v", fmt.Errorf("Couldn't parse date, %w", err))
		log.Fatalln("Error in TodaysRowsCsvRecordPredicate:", err)
	}
	// return startedTs.Truncate(time.Hour).Equal(time.Now().Truncate(time.Hour))
	return startedTs.Format(config.DefaultDatePattern) == time.Now().Format(config.DefaultDatePattern)
}

//NewCsvFile creates a new CsvFile with the given path and default header
func NewCsvFile(path string) CsvFile {
	return CsvFile{Path: path, Header: GetCsvHeader(), Records: []CsvRecord{}}
}

//AddRecord adds (appends) a given record
func (f *CsvFile) AddRecord(rec CsvRecord) {
	rec._idx = len(f.Records)
	f.Records = append(f.Records, rec)
}

//UpdateRecord replaces record at the given index with the new record.
func (f *CsvFile) UpdateRecord(idx int, rec CsvRecord) error {
	if idx < 0 || idx > len(f.Records) {
		return errors.New("Index is out of range")
	}
	rec._idx = idx
	f.Records[idx] = rec
	log.Println("Updated CSV Record at index", idx)
	return nil
}

//ReadAll reads CSV file from disk with all records
func (f *CsvFile) ReadAll() {
	f.Read(AllRowsCsvRecordPredicate)
}

//ReadAll reads CSV file from disk with records which satisfy recordFilter predicate.
//Records, not matching the predicate will not be added to CsvFile.Records
func (f *CsvFile) Read(recordFilter CsvRecordPredicate) {
	fcsv, err := os.Open(f.Path)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(fcsv)
	//Read header to skip it. Maybe later add a param like "readHeader: bool"
	header, err := reader.Read()
	if err != nil {
		log.Println("Error reading header:", err)
	}
	f.Header = header
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV records: %v", err)
	}
	countFiltered := 0
	for idx, rec := range records {
		csvRec, _ := NewCsvRecord(rec) //ignore validation errors when reading from file
		csvRec._idx = idx
		if recordFilter(csvRec) {
			f.AddRecord(csvRec)
			countFiltered++
		}
	}
	log.Printf("Read (filtered) %v rows out of total %v in %q\n", countFiltered, len(records), f.Path)
}

func (f *CsvFile) Write() {
	fcsv, err := os.Create(f.Path)
	if err != nil {
		log.Fatal(err)
	}
	writer := csv.NewWriter(fcsv)
	writer.Comma = ','
	err = writer.Write(f.Header)
	err = writer.WriteAll(f.Records.AsRows())
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()
	log.Printf("Flushed %v data rows in %v", len(f.Records), f.Path)
}
