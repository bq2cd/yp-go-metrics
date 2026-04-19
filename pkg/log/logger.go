package log

const (
	_levelDisabled Level = iota
	// LevelFatal corresponds to the most severe events, causing the program to terminate.
	LevelFatal
	// LevelError corresponds to events caused by errors in the program.
	LevelError
	// LevelInfo corresponds to informational events.
	LevelInfo
	// LevelDebug corresponds to events aimed at debugging program's implementation.
	LevelDebug
)

// Level defines a logging level as an integer.
type Level int8

// Logger defines high-level API to interact with underlying logging implementation.
type Logger interface {
	Fatal() EventBuilder
	Error() EventBuilder
	Info() EventBuilder
	Debug() EventBuilder

	With(fields ...Field) Logger
	WithErr(err error) Logger
	Sync() error
}

// loggerImpl is an internal abstraction over logging implementation
type loggerImpl interface {
	clone() loggerImpl
	log(lvl Level, msg string, fields ...Field)
	with(fields ...Field) loggerImpl
	sync() error
}

type baseLogger struct {
	impl loggerImpl
}

// Fatal returns new log event builder with log level set to [LevelFatal].
func (l *baseLogger) Fatal() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelFatal)
}

// Error returns new log event builder with log level set to [LevelError].
func (l *baseLogger) Error() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelError)
}

// Info returns new log event builder with log level set to [LevelInfo].
func (l *baseLogger) Info() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelInfo)
}

// Debug returns new log event builder with log level set to [LevelDebug].
func (l *baseLogger) Debug() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelDebug)
}

// With returns new instance of a logger with provided fields.
func (l *baseLogger) With(fields ...Field) Logger {
	return &baseLogger{
		impl: l.impl.with(fields...),
	}
}

// WithErr returns new instance of a logger with provider error.
func (l *baseLogger) WithErr(err error) Logger {
	return l.With(Err(errorDefaultKey, err))
}

// Sync flushes any pending log events to the wire (writes them to the configured destination).
func (l *baseLogger) Sync() error {
	return l.impl.sync()
}
