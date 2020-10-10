package log

import (
	"fmt"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/logrusorgru/aurora"
)

type level string

const (
	LevelDebug   level = "DEBUG"
	LevelInfo    level = "INFO"
	LevelWarning level = "WARNING"
	LevelStats   level = "STATS"
	LevelError   level = "ERROR"
	LevelFatal   level = "FATAL"
)

type Log interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Stats(data interface{})
	ErrorOn(err interface{})
	FatalOn(err interface{})
}

type log struct {
	l func(*Entry)
}

func New(f ...func(*Entry)) Log {
	PanicIf(len(f) > 1, "f must be nil or one func")
	l := defaultLog
	if len(f) == 1 {
		l = f[0]
	}
	return &log{
		l: l,
	}
}

type Entry struct {
	Time       time.Time   `json:"time"`
	Level      level       `json:"level"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	StackTrace string      `json:"stackTrace"`
}

func color(level level) func(interface{}) aurora.Value {
	switch level {
	case LevelDebug:
		return aurora.BgMagenta
	case LevelInfo:
		return aurora.Cyan
	case LevelWarning:
		return aurora.Yellow
	case LevelError:
		return aurora.BrightRed
	case LevelStats:
		return aurora.Green
	case LevelFatal:
		return aurora.BrightMagenta
	default:
		return aurora.White
	}
}

func defaultLog(e *Entry) {
	fmtStr := "%s\t%s"
	fullArgs := make([]interface{}, 0, 5)
	fullArgs = append(fullArgs, e.Time.Format("2006-01-02 15:04:05.000"), e.Level)
	if e.Message != "" {
		fmtStr += "\t%s"
		fullArgs = append(fullArgs, e.Message)
	}
	if e.Data != nil {
		if _, ok := e.Data.(fmt.Stringer); ok {
			fmtStr += "\t%s"
		} else {
			fmtStr += "\t%#v"
		}
		fullArgs = append(fullArgs, e.Data)
	}
	if e.StackTrace != "" {
		fmtStr += "\n%s"
		fullArgs = append(fullArgs, e.StackTrace)
	}
	Println(aurora.Sprintf(color(e.Level)(fmtStr), fullArgs...))
}

func (l *log) message(level level, f string, args ...interface{}) {
	l.l(&Entry{
		Time:    Now(),
		Level:   level,
		Message: Strf(f, args...),
	})
}

func (l *log) Debug(f string, args ...interface{}) {
	l.message(LevelDebug, f, args...)
}

func (l *log) Info(f string, args ...interface{}) {
	l.message(LevelInfo, f, args...)
}

func (l *log) Warning(f string, args ...interface{}) {
	l.message(LevelWarning, f, args...)
}

func (l *log) Stats(data interface{}) {
	l.l(&Entry{
		Time:  Now(),
		Level: LevelStats,
		Data:  data,
	})
}

func (l *log) error(level level, i interface{}) {
	if err := ToError(i); err != nil {
		l.l(&Entry{
			Time:       Now(),
			Level:      level,
			Message:    err.Message(),
			StackTrace: err.StackTrace(),
		})
	}
}

func (l *log) ErrorOn(err interface{}) {
	l.error(LevelError, err)
}

func (l *log) FatalOn(err interface{}) {
	l.error(LevelFatal, err)
	ExitOn(err)
}
