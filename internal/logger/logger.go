package logger

import (
	// external logger library
	"os"

	log "github.com/sirupsen/logrus"
)

var Log *log.Logger

func Setup(verbose bool) *log.Logger {
	Log = log.New()
	Log.SetOutput(os.Stdout)
	Log.Formatter = &log.JSONFormatter{}

	if verbose {
		Log.Level = log.DebugLevel
		Log.Debug("Logging verbosely")
	} else {
		Log.Level = log.InfoLevel
	}

	return Log
}
