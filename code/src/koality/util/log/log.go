package log

import (
	"fmt"
	"github.com/scale-it/go-log"
	"log/syslog"
)

var (
	// TODO(andrey) change to where we want the log files to go
	fileName    = "log.txt"
	formatter   = log.StdFormatter{"[root]", log.Lmicroseconds | log.Lshortfile, false}
	maxFileSize = 1 << 20
)

var Logger *log.Logger

func init() {
	Logger = log.New()

	// TODO(andrey) change myprog to deployment address or license key
	logwriter, _ := syslog.New(syslog.LOG_NOTICE, "my_program")

	Logger.AddHandler(logwriter, log.Levels.Trace, formatter)

	logFile, _ := log.NewRotFile(fileName, false, maxFileSize, 0)

	Logger.AddHandler(&logFile, log.Levels.Trace, formatter)
}

func Debug(v ...interface{}) {
	Logger.Log(log.Levels.Debug, fmt.Sprintln(v...))
}

func Info(v ...interface{}) {
	Logger.Log(log.Levels.Info, fmt.Sprintln(v...))
}

func Warning(v ...interface{}) {
	Logger.Log(log.Levels.Warning, fmt.Sprintln(v...))
}

func Error(v ...interface{}) {
	Logger.Log(log.Levels.Error, fmt.Sprintln(v...))
}

func Critical(v ...interface{}) {
	Logger.Log(log.Levels.Critical, fmt.Sprintln(v...))
}
