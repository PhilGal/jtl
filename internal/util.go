package cmd

import (
	"fmt"
	"log"
)

//LogFatal logs an error message, and optionally prints it into a standard output
func LogFatal(message string, isStdout bool) {
	if isStdout {
		fmt.Println(message)
	}
	log.Fatalln(message)
}
