package log

const (
	_levelDisabled Level = iota
	LevelFatal
	LevelError
	LevelInfo
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
	Sync()
}

// loggerImpl is an internal abstraction over logging implementation
type loggerImpl interface {
	clone() loggerImpl
	log(lvl Level, msg string, fields ...Field)
	with(fields ...Field) loggerImpl
	sync()
}

type baseLogger struct {
	impl loggerImpl
}

func (l *baseLogger) Fatal() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelFatal)
}

func (l *baseLogger) Error() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelError)
}

func (l *baseLogger) Info() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelInfo)
}

func (l *baseLogger) Debug() EventBuilder {
	return newEventBuilder(l.impl.clone(), LevelDebug)
}

func (l *baseLogger) With(fields ...Field) Logger {
	return &baseLogger{
		impl: l.impl.with(fields...),
	}
}

func (l *baseLogger) Sync() {
	l.impl.sync()
}
