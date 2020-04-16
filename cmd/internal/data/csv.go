package data

import (
	"encoding/csv"
	"log"
	"os"
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
	ID        string
	StartedTs string
	Comment   string
	TimeSpent string `validate:"regexp="^(\d+)[dhm]$"`
	Ticket    string `validate:"regexp="(([A-Za-z]{1,10})-?)[A-Z]+-\d+"`
	Category  string
}

//AsRow represents a CSV-writable row
func (r *CsvRecord) AsRow() []string {
	return []string{r.ID, r.StartedTs, r.Comment, r.TimeSpent, r.Ticket, r.Category}
}

//NewCsvRecord created a new CsvRecord from a slice representing a single CSV row.
func NewCsvRecord(rec []string) CsvRecord {
	numberOfFieldsInCsvRecord := 6
	if len(rec) != numberOfFieldsInCsvRecord {
		log.Fatalf("Cannot create CsvRecord. Slice size is %v, expected: %v", len(rec), numberOfFieldsInCsvRecord)
	}
	return CsvRecord{
		ID:        rec[0],
		StartedTs: rec[1],
		Comment:   rec[2],
		TimeSpent: rec[3],
		Ticket:    rec[4],
		Category:  rec[5],
	}
}

//CsvRecords is a wrapper on []CsvRecord
type CsvRecords []CsvRecord

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

//CsvFile represents a CSV file with header and records. Use Read() and Write() to read and write data from/to disk.
type CsvFile struct {
	Path    string
	Header  CsvHeader
	Records CsvRecords
}

type CsvRecordPredicate func(CsvRecord) bool

var AllRowsCsvRecordPredicate = func(r CsvRecord) bool {
	return true
}

var RowsWithoutIDsCsvRecordPredicate = func(r CsvRecord) bool {
	return r.ID == ""
}

//NewCsvFile creates a new CsvFile with the given path and default header
func NewCsvFile(path string) CsvFile {
	return CsvFile{Path: path, Header: GetCsvHeader(), Records: []CsvRecord{}}
}

//AddRecord adds (appends) a given record
func (f *CsvFile) AddRecord(rec CsvRecord) {
	f.Records = append(f.Records, rec)
}

//ReadAll reads CSV file from disk with all records
func (f *CsvFile) ReadAll() {
	f.Read(AllRowsCsvRecordPredicate)
}

//ReadAll reads CSV file from disk with records which satisfy recordFilter predicate.
//Records, not matching the predicate will not be added to CsvFile.Records.
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
	for _, rec := range records {
		csvRec := NewCsvRecord(rec)
		if recordFilter(csvRec) {
			f.AddRecord(csvRec)
		}
	}
}

func (f *CsvFile) Write() {
	fcsv, err := os.Create(f.Path)
	if err != nil {
		log.Fatal(err)
	}
	// f.Read()
	// f.AddRecord(NewCsvRecord(row))
	writer := csv.NewWriter(fcsv)
	writer.Comma = ','
	err = writer.Write(f.Header)
	err = writer.WriteAll(f.Records.AsRows())
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()
}
