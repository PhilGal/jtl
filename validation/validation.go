package validation

import (
	"fmt"
	"github.com/philgal/jtl/util"
	"log"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
)

//Validate is an entry point for validation
var Validate *validator.Validate

const (
	jiraTicketRegexp     = `(([A-Za-z]{1,10})-?)[A-Z]+-\d+`
	timeSpentRegexp      = `^(\d+d)? ?(\d+h)? ?(\d+m)?$`
	maxDurationInMinutes = 360
)

//InitValidator initializes validation and registers custom validators
func InitValidator() {
	Validate = validator.New()
	Validate.RegisterValidation("jiraticket", validateJiraTicketName)
	Validate.RegisterValidation("timespent", validateTimeSpent)
}

func validateJiraTicketName(fl validator.FieldLevel) bool {
	return validateRegexp(convertFieldToString(fl), jiraTicketRegexp, "Error validating Jira ticket name")
}

func validateTimeSpent(fl validator.FieldLevel) bool {
	isFormatValid := validateRegexp(convertFieldToString(fl), timeSpentRegexp, "Error validating timeSpent")
	if !isFormatValid {
		return false
	}
	durationMinutes, err := util.DurationToMinutes(fl.Field().String())
	handleError(err, "")
	if durationMinutes >= maxDurationInMinutes {
		fmt.Println(fmt.Errorf(
			"time spent on one ticket must not exceed %d minutes\n"+
				"Please add split a ticket with some break gap ;)", maxDurationInMinutes).Error())
		return false
	}
	return true
}

func validateRegexp(val string, regexpPattern string, message string) bool {
	match, err := regexp.MatchString(regexpPattern, val)
	handleError(err, message)
	return match
}

func convertFieldToString(fl validator.FieldLevel) string {
	return strings.TrimSpace(fl.Field().String())
}

func handleError(err error, message string) {
	if err != nil {
		fmt.Println(err)
		log.Fatalf("%v: %v\n", message, err)
	}
}
