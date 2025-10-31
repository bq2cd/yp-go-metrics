package log

import "sync"

type EventBuilder interface {
	EventFieldAdder

	With(fields ...Field) EventBuilder
	WithErr(err error) EventBuilder
	Msg(msg string)
	Send()
}

type eventBuilder struct {
	mu     sync.RWMutex
	logger loggerImpl
	level  Level
	fields FieldSet
}

func newEventBuilder(logger loggerImpl, level Level) *eventBuilder {
	return &eventBuilder{
		logger: logger,
		level:  level,
		fields: make(FieldSet, 0, 8), // 8 fields should be enough for an average log entry
	}
}

func (e *eventBuilder) addField(field Field) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.fields = append(e.fields, field)
}

// With appends given fields to the current log event.
func (e *eventBuilder) With(fields ...Field) EventBuilder {
	e.mu.RLock()
	defer e.mu.RUnlock()
	newCap := cap(e.fields)
	if len(e.fields)+len(fields) >= newCap {
		newCap += len(fields) * 2
	}
	newFields := make(FieldSet, 0, newCap)

	return &eventBuilder{
		logger: e.logger,
		level:  e.level,
		fields: append(append(newFields, e.fields...), fields...),
	}
}

// WithErr appends provided error under `error` field.
// If a custom field name is needed, use [Err] method.
func (e *eventBuilder) WithErr(err error) EventBuilder {
	return e.Err(errorDefaultKey, err)
}

// Msg sends current log event to the underlying logging implementation.
// The event itself is not discarded and can be reused.
func (e *eventBuilder) Msg(msg string) {
	if e.level == _levelDisabled {
		return
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.logger.log(e.level, msg, e.fields...)
}

// Send is similar to `Msg` method except it sends empty message field.
// This method is provided for convenience.
func (e *eventBuilder) Send() {
	// we duplicate the logic here to maintain the same level of the call stack
	if e.level == _levelDisabled {
		return
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.logger.log(e.level, "", e.fields...)
}
