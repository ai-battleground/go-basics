package service

import (
	"fmt"
	"log"
	"time"
)

const LogTimestampFormat = "2006-01-02 15:04:05.999" // prettier than RFC3339Nano

type Logger struct {
	IsDebug         bool
	Prefix          string
	timestampFormat string
}

func (l Logger) Info(msg string) {
	l.emit(msg, "INFO")
}

func (l Logger) Infof(msg string, parms ...interface{}) {
	l.Info(fmt.Sprintf(msg, parms...))
}

func (l Logger) Debug(msg string) {
	if l.IsDebug {
		l.emit(msg, "DEBUG")
	}
}

func (l Logger) Debugf(msg string, parms ...interface{}) {
	l.Debug(fmt.Sprintf(msg, parms...))
}

func (l Logger) emit(msg string, level string) {
	log.Printf("%s %s %s %s\n", time.Now().Format(LogTimestampFormat), level, l.Prefix, msg)
}

func (l Logger) WithPrefix(prefix string) Logger {
	return Logger{
		Prefix:  prefix,
		IsDebug: l.IsDebug,
	}
}

func (l Logger) PrependPrefix(prefix string) Logger {
	return l.WithPrefix(fmt.Sprintf("%s%s", prefix, l.Prefix))
}
