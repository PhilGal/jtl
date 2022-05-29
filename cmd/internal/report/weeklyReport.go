package report

//WeeklyReport represents a summary of tickets logged in a week, including tracked hours and number of pushed to Jira
type WeeklyReport struct {
	weekStart    string
	weekEnd      string
	totalTasks   int //including aliased, len(records)
	pushedTasks  int //tasks with ids
	totalMinutes int
	_dueHours    int //from configuration, how many ours it is required to log
	hoursLeft    int //number dueHours - totalHours, convert to human readable time
	// aliasReports []aliasReport
}

func (wr *WeeklyReport) totalTime() string {
	return minutesToDurationString(wr.totalMinutes)
}
