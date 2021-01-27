package parser

import "log"

type logger struct {
	verbose bool
}

func newLogger(verbose bool) *logger {
	return &logger{
		verbose: verbose,
	}
}

// Logf does what log.Printf does.
func (l *logger) Logf(format string, vv ...interface{}) {
	log.Printf(format, vv...)
}

// Debugf does what log.Printf does only if verbose is set to true.
func (l *logger) Debugf(format string, vv ...interface{}) {
	if l.verbose {
		log.Printf(format, vv...)
	}
}
