package hiweb

import (
	"fmt"
	"strings"
)

type Logger interface {
	Error(f interface{}, v ...interface{})
	Debug(f interface{}, v ...interface{})
	Warning(f interface{}, v ...interface{})
	Info(f interface{}, v ...interface{})
}

type DefaultLogger struct {
}

func (dl *DefaultLogger) Error(f interface{}, v ...interface{}) {
	formatLog(f, v...)
}

func (dl *DefaultLogger) Warning(f interface{}, v ...interface{}) {
	formatLog(f, v...)
}

func (dl *DefaultLogger) Info(f interface{}, v ...interface{}) {
	formatLog(f, v...)
}

func (dl *DefaultLogger) Debug(f interface{}, v ...interface{}) {
	formatLog(f, v...)
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
