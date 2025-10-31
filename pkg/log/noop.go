package log

// NewNoopLogger returns an instance of no-op logger
// which discards all messages thrown at it.
func NewNoopLogger() Logger {
	return &baseLogger{
		impl: &noopLogger{},
	}
}

type noopLogger struct{}

func (l *noopLogger) log(lvl Level, msg string, fields ...Field) {}
func (l *noopLogger) clone() loggerImpl                          { return l }
func (l *noopLogger) with(fields ...Field) loggerImpl            { return l }
func (l *noopLogger) sync() error                                { return nil }
