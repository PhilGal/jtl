package report

type WeeklyReport struct {
	weekStart    string
	weekEnd      string
	totalTasks   int //including aliased, len(records)
	totalMinutes int
	_dueHours    int //from configuration, how many ours it is required to log
	hoursLeft    int //number dueHours - totalHours, convert to human readable time
	// aliasReports []aliasReport
}

func (wr *WeeklyReport) totalTime() string {
	return minutesToDurationString(wr.totalMinutes)
}
