package csv

import (
	ecsv "encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/philgal/jtl/cmd/internal/config"
)

// File represents a CSV file with header and records. Use Read() and Write() to read and write data from/to disk
type File struct {
	Path    string
	Header  []string
	Records []Record
}

// CsvRecords is a wrapper on []CsvRecord
// type CsvRecords []Record

// Record represents a single record in a CSV file
type Record struct {
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
func (r Record) GetIdx() int {
	return r._idx
}

// AsRow represents a CSV-writable row
func (r Record) AsRow() []string {
	return []string{r.ID, r.StartedTs, r.Comment, r.TimeSpent, r.Ticket}
}

// IsPushed returns true is the item has been already pushed to Jira server
func (r Record) IsPushed() bool {
	return r.ID != ""
}

// TodaysRowsCsvRecordPredicate filters rows with startedTs = today
var TodaysRowsCsvRecordPredicate = func(r Record) bool {
	startedTs, err := time.Parse(config.DefaultDateTimePattern, r.StartedTs)
	if err != nil {
		fmt.Printf("Error! %v", fmt.Errorf("couldn't parse date, %w", err))
		log.Fatalln("Error in TodaysRowsCsvRecordPredicate:", err)
	}
	// return startedTs.Truncate(time.Hour).Equal(time.Now().Truncate(time.Hour))
	return startedTs.Format(config.DefaultDatePattern) == time.Now().Format(config.DefaultDatePattern)
}

// NewCsvFile creates a new File with the given path and default header
func NewCsvFile(path string) File {
	return File{Path: path, Header: config.Header(), Records: []Record{}}
}

// AddRecord adds (appends) a given record
func (f *File) AddRecord(rec Record) {
	rec._idx = len(f.Records)
	f.Records = append(f.Records, rec)
}

// UpdateRecord replaces record at the given index with the new record.
func (f *File) UpdateRecord(rec Record) {
	f.Records[rec._idx] = rec
	log.Printf("Updated CSV Record: %v\n", rec)
}

// ReadAll reads CSV file from disk with all records
func (f *File) ReadAll() {
	f.Read(func(cr Record) bool { return true })
}

// ReadAll reads CSV file from disk with records which satisfy recordFilter predicate.
// Records, not matching the predicate will not be added to File.Records
func (f *File) Read(recordFilter func(Record) bool) {
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

func (f *File) Write() {
	fcsv, err := os.Create(f.Path)
	if err != nil {
		log.Fatal(err)
	}
	writer := ecsv.NewWriter(fcsv)
	writer.Comma = ','
	hErr := writer.Write(f.Header)
	for _, r := range f.Records {
		row := r.AsRow()
		if len(row) != len(f.Header) {
			errors.Join(err, errors.New(fmt.Sprintf("Row %s doesn't the header %s\n", row, f.Header)))
		}
		rErr := writer.Write(row)
		if rErr != nil {
			err = errors.Join(err, rErr)
			log.Printf("Error writing record %v: %s\n", r, rErr)
		}
	}
	err = errors.Join(hErr, err)
	if err != nil {
		panic(err)
	}
	writer.Flush()
	log.Printf("Flushed %v data rows in %v", len(f.Records), f.Path)
}

func Filter(recs []Record, predicate func(Record) bool) []Record {
	var filteredRecs []Record
	for _, r := range recs {
		if predicate(r) {
			filteredRecs = append(filteredRecs, r)
		}
	}
	return filteredRecs
}

func (f *File) Filter(predicate func(Record) bool) []Record {
	return Filter(f.Records, predicate)
}

func newCsvRec(rec []string) Record {
	return Record{
		ID:        rec[0],
		StartedTs: rec[1],
		Comment:   rec[2],
		TimeSpent: rec[3],
		Ticket:    rec[4],
	}
}
