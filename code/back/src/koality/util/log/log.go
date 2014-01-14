package log

import (
	"fmt"
	"github.com/scale-it/go-log"
	"log/syslog"
)

var Logger *log.Logger

func init() {
	Logger = log.New()

	formatter := log.StdFormatter{"[root]", log.Lmicroseconds | log.Lshortfile, false}

	// TODO(andrey) change myprog to deployment address or license key
	logwriter, _ := syslog.New(syslog.LOG_NOTICE, "my_program")

	Logger.AddHandler(logwriter, log.Levels.Trace, formatter)
}

func Debug(v ...interface{}) {
	Logger.Log(log.Levels.Debug, fmt.Sprintln(v...))
}

func Debugf(format string, v ...interface{}) {
	Logger.Log(log.Levels.Debug, fmt.Sprintf(format+"\n", v...))
}

func Info(v ...interface{}) {
	Logger.Log(log.Levels.Info, fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	Logger.Log(log.Levels.Info, fmt.Sprintf(format+"\n", v...))
}

func Warning(v ...interface{}) {
	Logger.Log(log.Levels.Warning, fmt.Sprintln(v...))
}

func Warningf(format string, v ...interface{}) {
	Logger.Log(log.Levels.Warning, fmt.Sprintf(format+"\n", v...))
}

func Error(v ...interface{}) {
	Logger.Log(log.Levels.Error, fmt.Sprintln(v...))
}

func Errorf(format string, v ...interface{}) {
	Logger.Log(log.Levels.Error, fmt.Sprintf(format+"\n", v...))
}

func Critical(v ...interface{}) {
	Logger.Log(log.Levels.Critical, fmt.Sprintln(v...))
}

func Criticalf(format string, v ...interface{}) {
	Logger.Log(log.Levels.Critical, fmt.Sprintf(format+"\n", v...))
}
