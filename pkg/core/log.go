package core

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/logrusorgru/aurora"
)

var (
	theLog        *log
	logConfigured = false
	fmtStr        = "%s\t%s\t%s"
	nowLogFormat  = "2006-01-02 15:04:05.000"
)

func init() {
	ConfigLog(nil, false, true)
	logConfigured = false
}

type Log interface {
	Debug(f string, args ...interface{})
	Info(f string, args ...interface{})
	Warning(f string, args ...interface{})
	ErrorOn(err error)
	FatalOn(err error)
}

func GetLog() Log {
	return theLog
}

// ConfigLog may only be called once to set the log behaviour
// for the application and get the log instance
func ConfigLog(l func(string), asJSON, withColors bool) {
	PanicIf(logConfigured, "log may only be configured once")
	logConfigured = true
	if l == nil {
		l = func(s string) { fmt.Println(s) }
	}
	theLog = &log{
		l:          l,
		asJSON:     asJSON,
		withColors: withColors,
	}
}

type log struct {
	l          func(string)
	asJSON     bool
	withColors bool
}

func (l *log) log(color func(interface{}) aurora.Value, lev, f string, args ...interface{}) {
	if l.asJSON {
		j, _ := json.Marshal(map[string]interface{}{
			"time":    Now(),
			"level":   lev,
			"message": fmt.Sprintf(f, args...),
		})
		if l.withColors {
			l.l(aurora.Sprintf(color(string(j))))
		} else {
			l.l(string(j))
		}
	} else {
		if l.withColors {
			l.l(aurora.Sprintf(color(fmtStr), Now().Format(nowLogFormat), lev, fmt.Sprintf(f, args...)))
		} else {
			l.l(fmt.Sprintf(fmtStr, Now().Format(nowLogFormat), lev, fmt.Sprintf(f, args...)))
		}
	}
}

func (l *log) error(color func(interface{}) aurora.Value, lev string, err error) {
	if err != nil {
		l.log(color, lev, "%s\n%s", err.Error(), string(debug.Stack()))
	}
}

func (l *log) Debug(f string, args ...interface{}) {
	l.log(aurora.BrightBlack, "DEBUG", f, args...)
}

func (l *log) Info(f string, args ...interface{}) {
	l.log(aurora.Cyan, "INFO", f, args...)
}

func (l *log) Warning(f string, args ...interface{}) {
	l.log(aurora.Yellow, "WARNING", f, args...)
}

func (l *log) ErrorOn(err error) {
	l.error(aurora.BrightRed, "ERROR", err)
}

func (l *log) FatalOn(err error) {
	l.error(aurora.BrightMagenta, "FATAL", err)
	if err != nil {
		os.Exit(2)
	}
}
