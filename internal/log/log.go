package log

import (
	"fmt"
	"math"
	"os"
	"slices"
	"time"

	"github.com/philgal/jtl/internal/config"
	"github.com/philgal/jtl/internal/csv"
	"github.com/philgal/jtl/internal/duration"
)

type Executor interface {
	Execute()
}

type ExecutorArgs struct {
	Ticket    string
	TimeSpent string
	Comment   string
	StartedTs string
}

type AutoFitting struct {
	ExecutorArgs
}

type Normal struct {
	ExecutorArgs
}

func (e Normal) Execute() {
	file := csv.ReadFile(config.DataFilePath())
	if duration.ToMinutes(e.TimeSpent) <= 0 {
		fmt.Println("Time spent must be greater than 0")
		os.Exit(1)
	}
	file.AddRecord(csv.Record{
		ID:        "",
		StartedTs: e.StartedTs,
		Comment:   e.Comment,
		TimeSpent: e.TimeSpent,
		Ticket:    e.Ticket,
	})
	file.Write()
}

func (e AutoFitting) Execute() {
	file := csv.ReadFile(config.DataFilePath())

	// calculate how much is left of the startedTs
	slices.SortFunc(file.Records, func(a csv.Record, b csv.Record) int {
		return duration.ParseTimeTruncatedToDate(a.StartedTs).Compare(duration.ParseTimeTruncatedToDate(b.StartedTs))
	})

	logDate, _ := time.Parse(config.DefaultDateTimePattern, e.StartedTs)
	//if dates are equal, count hours
	sameDateRecs := file.Filter(csv.SameDateRecordsFilter(logDate))
	minutesSpentToDate := timeSpentToDateInMin(sameDateRecs, logDate)

	// for example 500 > 480 -> 20m to log
	// todo: make max daily duration configurable
	if duration.ToMinutes(e.TimeSpent)+minutesSpentToDate >= duration.EightHoursInMin {
		// todo: make this logic optional if some "distribute" flag is set
		adjustableRecords := csv.Filter(sameDateRecs, func(r csv.Record) bool { return r.ID == "" })
		if len(adjustableRecords) == 0 {
			fmt.Printf("You have already logged %s, will not log more\n", duration.ToString(minutesSpentToDate))
			return
		}

		// calc timeSpent for each record
		totalTimeSpentToLog := math.Min(float64(minutesSpentToDate+duration.ToMinutes(e.TimeSpent)), duration.EightHoursInMin)
		totalRecordsToLog := len(sameDateRecs) + 1
		timeSpentPerRec := int(totalTimeSpentToLog / float64(totalRecordsToLog))
		e.TimeSpent = duration.ToString(timeSpentPerRec)

		for idx, r := range adjustableRecords {
			// calc startedTs for each record
			if idx > 0 {
				prevRec := adjustableRecords[idx-1]
				startedTs, _ := time.Parse(config.DefaultDateTimePattern, prevRec.StartedTs)
				startedTs = startedTs.Add(time.Duration(timeSpentPerRec) * time.Minute)
				r.StartedTs = startedTs.Format(config.DefaultDateTimePattern)
			}
			// set the values
			r.TimeSpent = duration.ToString(timeSpentPerRec)
			file.UpdateRecord(r)
		}
	} else {
		// we have some time to log: do dynamic timeSpent & startedTs calculation for all today's records
		// if today's records are empty, fill 8h with multiple records of 4h.
		// if today's records are not empty, calculate timeSpent and startedTs based on the existing records
		timeSpentMin := int(math.Min(float64(duration.EightHoursInMin-minutesSpentToDate), duration.EightHoursInMin/2))
		e.TimeSpent = duration.ToString(timeSpentMin)
		fmt.Printf("Time spent will is trimmed to %s, to not to exceed %s\n",
			e.TimeSpent,
			duration.ToString(duration.EightHoursInMin))
	}

	// adjust startedTs to the last record: new startedTs = last rec.StartedTs + calculated time spent
	sameDateRecs = file.Filter(csv.SameDateRecordsFilter(logDate))
	if l := len(sameDateRecs); l > 0 {
		lastRec := sameDateRecs[l-1]
		lastRecStaredAt := duration.ParseTime(lastRec.StartedTs)
		e.StartedTs = lastRecStaredAt.Add(time.Minute * time.Duration(duration.ToMinutes(lastRec.TimeSpent))).Format(config.DefaultDateTimePattern)
	}

	if e.TimeSpent != "0m" {
		file.AddRecord(csv.Record{
			ID:        "",
			StartedTs: e.StartedTs,
			Comment:   e.Comment,
			TimeSpent: e.TimeSpent,
			Ticket:    e.Ticket,
		})
		file.Write()
	} else {
		fmt.Println("Calculated time spent is 0m, will not log!")
		os.Exit(1)
	}
}

func timeSpentToDateInMin(sameDateRecs []csv.Record, logDate time.Time) int {
	var totalTimeSpentOnDate int
	for _, rec := range slices.Backward(sameDateRecs) {
		recDate, _ := time.Parse(config.DefaultDateTimePattern, rec.StartedTs)
		// logs files are already collected by months of the year, so it's enough to compare days
		if recDate.Day() == logDate.Day() {
			totalTimeSpentOnDate += duration.ToMinutes(rec.TimeSpent)
		} else {
			// items are sorted, so first occurrence of mismatched date means there will no more matches.
			return totalTimeSpentOnDate
		}
	}
	return totalTimeSpentOnDate
}

func sameDateRecords(recs []csv.Record, logDate time.Time) []csv.Record {
	return csv.Filter(recs, func(rec csv.Record) bool {
		return duration.ParseTimeTruncatedToDate(rec.StartedTs).Equal(logDate.Truncate(24 * time.Hour))
	})
}
