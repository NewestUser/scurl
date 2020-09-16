package scurl

import "fmt"

var mutedLogger = &logger{verbose: false}

type logger struct {
	verbose bool
}

func (l *logger) debug(a ...interface{}) {
	if l.verbose {
		fmt.Println(a...)
	}
}
