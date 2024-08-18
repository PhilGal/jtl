package csv

import (
	ecsv "encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/philgal/jtl/validation"
)

// CsvHeader represents header in a CSV file
type CsvHeader []string

var header = CsvHeader{"ID", "StartedTs", "Comment", "TimeSpent", "Ticket"}

// GetCsvHeader returns a CSV-header row
func GetCsvHeader() CsvHeader {
	return header
}

// CsvRec represents a single record in a CSV file
type CsvRec struct {
	_idx      int
	ID        string
	StartedTs string
	Comment   string
	// TimeSpent string `validate:"required"`
	// Ticket    string `validate:"required"`
	TimeSpent string `validate:"required,timespent"`
	Ticket    string `validate:"required,jiraticket"`
}

// GetIdx returns a row's index in CSV file
func (r *CsvRec) GetIdx() int {
	return r._idx
}

// AsRow represents a CSV-writable row
func (r *CsvRec) AsRow() []string {
	return []string{r.ID, r.StartedTs, r.Comment, r.TimeSpent, r.Ticket}
}

// IsPushed returns true is the item has been already pushed to Jira server
func (r *CsvRec) IsPushed() bool {
	return r.ID != ""
}

// CsvRecords is a wrapper on []CsvRecord
type CsvRecords []CsvRec

// Filter returns a copy of records, filtered using the given recordFilter
func (recs *CsvRecords) Filter(predicate func(CsvRec) bool) CsvRecords {
	filteredRecs := CsvRecords{}
	for _, r := range *recs {
		if predicate(r) {
			filteredRecs = append(filteredRecs, r)
		}
	}
	return filteredRecs
}

// AsRows converts CsvRecords into 2-d slice representing CSV {records X columns}
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

// CsvFile represents a CSV file with header and records. Use Read() and Write() to read and write data from/to disk
type CsvFile struct {
	Path    string
	Header  CsvHeader
	Records CsvRecords
}

// TodaysRowsCsvRecordPredicate filters rows with startedTs = today
var TodaysRowsCsvRecordPredicate = func(r CsvRec) bool {
	startedTs, err := time.Parse(config.DefaultDateTimePattern, r.StartedTs)
	if err != nil {
		fmt.Printf("Error! %v", fmt.Errorf("couldn't parse date, %w", err))
		log.Fatalln("Error in TodaysRowsCsvRecordPredicate:", err)
	}
	// return startedTs.Truncate(time.Hour).Equal(time.Now().Truncate(time.Hour))
	return startedTs.Format(config.DefaultDatePattern) == time.Now().Format(config.DefaultDatePattern)
}

// NewCsvFile creates a new CsvFile with the given path and default header
func NewCsvFile(path string) CsvFile {
	return CsvFile{Path: path, Header: GetCsvHeader(), Records: []CsvRec{}}
}

// AddRecord adds (appends) a given record
func (f *CsvFile) AddRecord(rec CsvRec) {
	if err := validation.Validate.Struct(rec); err != nil {
		fmt.Println("invalid record %w", err)
		panic(err)
	}
	rec._idx = len(f.Records)
	f.Records = append(f.Records, rec)
}

// UpdateRecord replaces record at the given index with the new record.
func (f *CsvFile) UpdateRecord(idx int, rec CsvRec) error {
	if idx < 0 || idx > len(f.Records) {
		return errors.New("index is out of range")
	}
	rec._idx = idx
	f.Records[idx] = rec
	log.Println("Updated CSV Record at index", idx)
	return nil
}

// ReadAll reads CSV file from disk with all records
func (f *CsvFile) ReadAll() {
	f.Read(func(cr CsvRec) bool { return true })
}

// ReadAll reads CSV file from disk with records which satisfy recordFilter predicate.
// Records, not matching the predicate will not be added to CsvFile.Records
func (f *CsvFile) Read(recordFilter func(CsvRec) bool) {
	fcsv, err := os.Open(f.Path)
	if err != nil {
		log.Fatal(err)
	}
	reader := ecsv.NewReader(fcsv)
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
		csvRec := newCsvRec(rec) //ignore validation errors when reading from file
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
	writer := ecsv.NewWriter(fcsv)
	writer.Comma = ','
	hErr := writer.Write(f.Header)
	rErr := writer.WriteAll(f.Records.AsRows())
	err = errors.Join(hErr, rErr)
	if err != nil {
		panic(err)
	}
	writer.Flush()
	log.Printf("Flushed %v data rows in %v", len(f.Records), f.Path)
}

func newCsvRec(rec []string) CsvRec {
	if len(rec) != len(header) {
		panic(fmt.Sprintf("Cannot create CsvRecord. Slice size is %v, expected: %v", len(rec), len(header)))
	}
	return CsvRec{
		ID:        rec[0],
		StartedTs: rec[1],
		Comment:   rec[2],
		TimeSpent: rec[3],
		Ticket:    rec[4],
	}
}
