package log

type eventBuilder struct {
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

func (e *eventBuilder) With(fields ...Field) EventBuilder {
	newCap := cap(e.fields)
	if len(e.fields)+len(fields) >= newCap {
		newCap += len(fields) * 2
	}

	newFields := make(FieldSet, 0, newCap)
	copy(newFields, e.fields)

	return &eventBuilder{
		logger: e.logger.clone(),
		level:  e.level,
		fields: append(newFields, fields...),
	}
}

func (e *eventBuilder) Msg(msg string) {
	if e.level == _levelDisabled {
		return
	}
	e.logger.log(e.level, msg, e.fields...)
}

func (e *eventBuilder) Send() {
	// we duplicate the logic here to maintain the same level of the call stack
	if e.level == _levelDisabled {
		return
	}
	e.logger.log(e.level, "", e.fields...)
}
