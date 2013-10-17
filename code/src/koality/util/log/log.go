package log

import (
	"github.com/scale-it/go-log"
	"log/syslog"
)

var fileName = "log.txt"
var formatter = log.StdFormatter{"[root]", log.Lmicroseconds | log.Lshortfile, false}
var maxFileSize = 1 << 20

var Logger *log.Logger

func init() {
	Logger = log.New()

	// TODO(andrey) change myprog to deployment address or license key
	logwriter, _ := syslog.New(syslog.LOG_NOTICE, "my_program")

	Logger.AddHandler(logwriter, log.Levels.Trace, formatter)

	logFile, _ := log.NewRotFile(fileName, false, maxFileSize, 0)

	Logger.AddHandler(&logFile, log.Levels.Trace, formatter)
}
