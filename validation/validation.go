package validation

import (
	"fmt"
	"log"
	"regexp"

	"github.com/go-playground/validator"
)

//Validate is an entry point for validation
var Validate *validator.Validate

const (
	jiraTicketRegexp = `(([A-Za-z]{1,10})-?)[A-Z]+-\d+`
	timeSpentRegexp  = `^(\d+)[dhm]$`
)

//InitValidator initializes validation and registers custom validators
func InitValidator() {
	fmt.Println("[DEBUG]", "Initializing validation")
	Validate = validator.New()
	Validate.RegisterValidation("jiraticket", validateJiraTicketName)
	Validate.RegisterValidation("timespent", validateTimeSpent)
}

func validateJiraTicketName(fl validator.FieldLevel) bool {
	return validateRegexp(fl, jiraTicketRegexp, "Error validating Jira ticket name")
}

func validateTimeSpent(fl validator.FieldLevel) bool {
	return validateRegexp(fl, timeSpentRegexp, "Error validating timeSpent")
}

func validateRegexp(fl validator.FieldLevel, regexpPattern string, message string) bool {
	val := fl.Field().String()
	match, err := regexp.MatchString(regexpPattern, val)
	if err != nil {
		log.Fatalf("%v: %v\n", message, err)
	}
	return match
}
