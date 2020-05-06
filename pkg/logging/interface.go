// Package logging //
package logging

// FlagshipLogConsumer consumes log messages produced by the logger
type FlagshipLogConsumer interface {
	Log(level LogLevel, message string, name string)
	SetLogLevel(logLevel LogLevel)
}

// FlagshipLogger produces log messages to be consumed by the log consumer
type FlagshipLogger interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string, err interface{})
}
