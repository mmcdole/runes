package util

import "fmt"

type LogLevel int

const (
	ErrorLogLevel LogLevel = iota
	InfoLogLevel
	DebugLogLevel
	TraceLogLevel
)

type Logger struct {
	LogLevel LogLevel
	Prefix   string
}

func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level > l.LogLevel {
		return
	}

	levelRunes := map[LogLevel]string{
		InfoLogLevel:  "info",
		ErrorLogLevel: "error",
		DebugLogLevel: "debug",
		TraceLogLevel: "trace",
	}[level]

	formattedMessage := fmt.Sprintf(message, args...)
	if l.Prefix != "" {
		fmt.Printf("[%s] [%s] %s\n", levelRunes, l.Prefix, formattedMessage)
	} else {
		fmt.Printf("[%s] %s\n", levelRunes, formattedMessage)
	}
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DebugLogLevel, message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ErrorLogLevel, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(InfoLogLevel, message, args...)
}

func (l *Logger) Trace(message string, args ...interface{}) {
	l.log(TraceLogLevel, message, args...)
}
