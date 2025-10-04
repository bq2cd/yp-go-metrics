package log

import (
	"maps"
	"sync"
)

// TestLogger extends regular Logger interface with extra method:
// `RecordedEvents()`.
// Its primary purpose is to be used in tests to record logging events.
type TestLogger interface {
	Logger
	RecordedEvents() []TestLogEvent
}

// TestLogEvent represents a logging event with essential
// attributes (level, message, extra fields).
// There are no timestamps or caller information as such are
// rarely needed in tests.
type TestLogEvent interface {
	Level() Level
	Message() string
	Fields() FieldSet
}

// NewTestLogger returns an instance of a special logger with the purpose of being used in tests.
// The logger will record all log events in-memory and allow to access
// them via `RecordedEvents()` method.
func NewTestLogger() TestLogger {
	recorder := &testLogEventRecorder{
		recorded: make([]TestLogEvent, 0, 16),
	}
	return &testLogger{
		baseLogger: &baseLogger{
			impl: &testLoggerImpl{
				fieldsByKey: make(map[string]Field, 8),
				logFunc:     recorder.recordEvent,
			},
		},
		recorder: recorder,
	}
}

type testLogger struct {
	*baseLogger
	recorder *testLogEventRecorder
}

// RecordedEvents returns a slice of the recorded log events.
func (l *testLogger) RecordedEvents() []TestLogEvent {
	return l.recorder.recordedEvents()
}

type testLogEvent struct {
	level   Level
	message string
	fields  FieldSet
}

// Level returns log level of the event.
func (e testLogEvent) Level() Level {
	return e.level
}

// Message returns message field of the event.
func (e testLogEvent) Message() string {
	return e.message
}

// Fields returns a slice of extra fields of the event.
func (e testLogEvent) Fields() FieldSet {
	return e.fields
}

type testLogEventRecorder struct {
	mu       sync.RWMutex
	recorded []TestLogEvent
}

func (r *testLogEventRecorder) recordEvent(lvl Level, msg string, fields ...Field) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = append(r.recorded, testLogEvent{
		level:   lvl,
		message: msg,
		fields:  fields,
	})
}

func (r *testLogEventRecorder) recordedEvents() []TestLogEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recorded[:]
}

type testLoggerImpl struct {
	mu          sync.RWMutex
	fieldsByKey map[string]Field
	logFunc     func(Level, string, ...Field)
}

func (l *testLoggerImpl) addFields(fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, f := range fields {
		l.fieldsByKey[f.Key] = f
	}
}

func (l *testLoggerImpl) log(lvl Level, msg string, fields ...Field) {
	l.mu.RLock()
	mergedFields := make(map[string]Field, len(l.fieldsByKey)+len(fields))
	maps.Copy(mergedFields, l.fieldsByKey)
	l.mu.RUnlock()
	for _, f := range fields {
		mergedFields[f.Key] = f
	}
	loggedFields := make([]Field, 0, len(mergedFields))
	for _, f := range mergedFields {
		loggedFields = append(loggedFields, f)
	}
	l.logFunc(lvl, msg, loggedFields...)
}

func (l *testLoggerImpl) clone() loggerImpl {
	l.mu.RLock()
	defer l.mu.RUnlock()
	newLogger := &testLoggerImpl{
		fieldsByKey: make(map[string]Field, len(l.fieldsByKey)*2),
		logFunc:     l.logFunc,
	}
	maps.Copy(newLogger.fieldsByKey, l.fieldsByKey)
	return newLogger
}

func (l *testLoggerImpl) with(fields ...Field) loggerImpl {
	newLogger := l.clone().(*testLoggerImpl)
	newLogger.addFields(fields...)
	return newLogger
}

func (l *testLoggerImpl) sync() error {
	return nil
}
