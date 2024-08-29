package report

//WeeklyReport represents a summary of tickets logged in a week, including tracked hours and number of pushed to Jira
type WeeklyReport struct {
	weekStart    string
	weekEnd      string
	totalTasks   int //including aliased, len(records)
	pushedTasks  int //tasks with ids
	totalMinutes int
}
