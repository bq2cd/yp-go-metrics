package log

import (
	"cmp"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestLogger(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(*testing.T, TestLogger)
	}{
		{
			name: "default",
			assertion: func(t *testing.T, got TestLogger) {
				assert.IsType(t, &testLogger{}, got)
				l := got.(*testLogger)
				assert.NotNil(t, l.recorder)
				assert.NotNil(t, l.recorder.recorded)
				assert.IsType(t, &testLoggerImpl{}, l.impl)
				li := l.impl.(*testLoggerImpl)
				assert.NotNil(t, li.fieldsByKey)
				assert.NotNil(t, li.logFunc)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, NewTestLogger())
		})
	}
}

func Test_testLogger_RecordedEvents(t *testing.T) {
	type fields struct {
		recorder *testLogEventRecorder
	}
	tests := []struct {
		name   string
		fields fields
		want   TestLogEventSet
	}{
		{
			name: "no events",
			fields: fields{
				recorder: &testLogEventRecorder{
					recorded: []TestLogEvent{},
				},
			},
			want: []TestLogEvent{},
		},
		{
			name: "single event",
			fields: fields{
				recorder: &testLogEventRecorder{
					recorded: []TestLogEvent{testLogEvent{level: LevelInfo, message: "an info message", fields: FieldSet{Int("i1", 123)}}},
				},
			},
			want: []TestLogEvent{testLogEvent{level: LevelInfo, message: "an info message", fields: FieldSet{Int("i1", 123)}}},
		},
		{
			name: "multiple events",
			fields: fields{
				recorder: &testLogEventRecorder{
					recorded: []TestLogEvent{
						testLogEvent{level: LevelInfo, message: "an info message", fields: FieldSet{Int("i1", 123)}},
						testLogEvent{level: LevelDebug, message: "a debug message", fields: FieldSet{Float("f1", 1.23)}},
					},
				},
			},
			want: []TestLogEvent{
				testLogEvent{level: LevelInfo, message: "an info message", fields: FieldSet{Int("i1", 123)}},
				testLogEvent{level: LevelDebug, message: "a debug message", fields: FieldSet{Float("f1", 1.23)}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &testLogger{
				baseLogger: &baseLogger{impl: &noopLogger{}},
				recorder:   tt.fields.recorder,
			}
			assert.Equal(t, tt.want, l.RecordedEvents())
		})
	}
}

func Test_testLogEvent_Level(t *testing.T) {
	type fields struct {
		level   Level
		message string
		fields  FieldSet
	}
	tests := []struct {
		name   string
		fields fields
		want   Level
	}{
		{
			name: "some level",
			fields: fields{
				level: LevelFatal,
			},
			want: LevelFatal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testLogEvent{
				level:   tt.fields.level,
				message: tt.fields.message,
				fields:  tt.fields.fields,
			}
			assert.Equal(t, tt.want, e.Level())
		})
	}
}

func Test_testLogEvent_Message(t *testing.T) {
	type fields struct {
		level   Level
		message string
		fields  FieldSet
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "some message",
			fields: fields{
				message: "not again",
			},
			want: "not again",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testLogEvent{
				level:   tt.fields.level,
				message: tt.fields.message,
				fields:  tt.fields.fields,
			}
			assert.Equal(t, tt.want, e.Message())
		})
	}
}

func Test_testLogEvent_Fields(t *testing.T) {
	type fields struct {
		level   Level
		message string
		fields  FieldSet
	}
	tests := []struct {
		name   string
		fields fields
		want   FieldSet
	}{
		{
			name: "some fields",
			fields: fields{
				fields: FieldSet{Int("i1", 123), Float("f1", 1.23)},
			},
			want: FieldSet{Int("i1", 123), Float("f1", 1.23)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testLogEvent{
				level:   tt.fields.level,
				message: tt.fields.message,
				fields:  tt.fields.fields,
			}
			assert.Equal(t, tt.want, e.Fields())
		})
	}
}

func Test_testLogEventRecorder_recordEvent(t *testing.T) {
	type fields struct {
		recorded []TestLogEvent
	}
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	type want struct {
		events []TestLogEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "empty fields",
			fields: fields{recorded: []TestLogEvent{}},
			args: args{
				lvl:    LevelInfo,
				msg:    "an operation",
				fields: FieldSet{},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "an operation",
						fields:  FieldSet{},
					},
				},
			},
		},
		{
			name:   "no pre-existing events",
			fields: fields{recorded: []TestLogEvent{}},
			args: args{
				lvl:    LevelError,
				msg:    "something went wrong",
				fields: FieldSet{Int("i1", 123), Err("error", errors.New("nasty error"))},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "something went wrong",
						fields:  FieldSet{Int("i1", 123), Err("error", errors.New("nasty error"))},
					},
				},
			},
		},
		{
			name: "some pre-existing events",
			fields: fields{recorded: []TestLogEvent{
				testLogEvent{
					level:   LevelInfo,
					message: "a normal operation",
					fields:  FieldSet{Int("i1", 456), Float("f1", -3.21)},
				},
			}},
			args: args{
				lvl:    LevelError,
				msg:    "something went wrong",
				fields: FieldSet{Int("i1", 123), Err("error", errors.New("nasty error"))},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "something went wrong",
						fields:  FieldSet{Int("i1", 123), Err("error", errors.New("nasty error"))},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "a normal operation",
						fields:  FieldSet{Int("i1", 456), Float("f1", -3.21)},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testLogEventRecorder{
				recorded: tt.fields.recorded,
			}
			r.recordEvent(tt.args.lvl, tt.args.msg, tt.args.fields...)
		})
	}
}

func Test_testLogEventRecorder_recordedEvents(t *testing.T) {
	type fields struct {
		recorded []TestLogEvent
	}
	tests := []struct {
		name   string
		fields fields
		want   []TestLogEvent
	}{
		// covered by `Test_testLogger_RecordedEvents`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testLogEventRecorder{
				recorded: tt.fields.recorded,
			}
			assert.Equal(t, tt.want, r.recordedEvents())
		})
	}
}

func Test_testLoggerImpl_addFields(t *testing.T) {
	type fields struct {
		fieldsByKey map[string]Field
		logFunc     func(Level, string, ...Field)
	}
	type args struct {
		fields []Field
	}
	type want struct {
		fieldsByKey map[string]Field
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "unique fields",
			fields: fields{
				fieldsByKey: map[string]Field{},
			},
			args: args{
				fields: FieldSet{Int("i1", 123), Str("s1", "bla1")},
			},
			want: want{
				fieldsByKey: map[string]Field{
					"i1": Int("i1", 123),
					"s1": Str("s1", "bla1"),
				},
			},
		},
		{
			name: "duplicate fields",
			fields: fields{
				fieldsByKey: map[string]Field{},
			},
			args: args{
				fields: FieldSet{Int("i1", 123), Str("s1", "bla1"), Int("i1", -123)},
			},
			want: want{
				fieldsByKey: map[string]Field{
					"i1": Int("i1", -123),
					"s1": Str("s1", "bla1"),
				},
			},
		},
		{
			name: "pre-existing fields",
			fields: fields{
				fieldsByKey: map[string]Field{
					"i1": Int("i1", -456),
					"b1": Bool("b1", true),
				},
			},
			args: args{
				fields: FieldSet{Int("i1", 123), Str("s1", "bla1"), Int("i1", -123)},
			},
			want: want{
				fieldsByKey: map[string]Field{
					"i1": Int("i1", -123),
					"s1": Str("s1", "bla1"),
					"b1": Bool("b1", true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &testLoggerImpl{
				fieldsByKey: tt.fields.fieldsByKey,
				logFunc:     tt.fields.logFunc,
			}
			l.addFields(tt.args.fields...)
			assert.Equal(t, tt.want.fieldsByKey, l.fieldsByKey)
		})
	}
}

func normaliseTestLogEvents(eventSlices ...[]TestLogEvent) {
	for _, events := range eventSlices {
		for _, event := range events {
			if e, ok := event.(testLogEvent); ok {
				if len(e.fields) == 0 {
					continue
				}
				slices.SortStableFunc(e.fields, func(a Field, b Field) int {
					return cmp.Compare(a.Key, b.Key)
				})
			}
		}
	}
}

func Test_testLoggerImpl_log(t *testing.T) {
	type fields struct {
		recorded    []TestLogEvent
		fieldsByKey map[string]Field
	}
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	type want struct {
		events []TestLogEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		args2  args
		want   want
		want2  want
	}{
		{
			name: "empty fields",
			fields: fields{
				recorded:    []TestLogEvent{},
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{},
					},
				},
			},
			args2: args{
				lvl:    LevelDebug,
				msg:    "2nd message",
				fields: FieldSet{},
			},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "2nd message",
						fields:  FieldSet{},
					},
				},
			},
		},
		{
			name: "multiple fields",
			fields: fields{
				recorded:    []TestLogEvent{},
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
					},
				},
			},
			args2: args{
				lvl:    LevelDebug,
				msg:    "2nd message",
				fields: FieldSet{Int("i2", 123), Dur("d2", 5*time.Second), Bool("b2", true)},
			},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "2nd message",
						fields:  FieldSet{Int("i2", 123), Dur("d2", 5*time.Second), Bool("b2", true)},
					},
				},
			},
		},
		{
			name: "multiple fields pre-existing",
			fields: fields{
				recorded: []TestLogEvent{},
				fieldsByKey: map[string]Field{
					"x1": Float("x1", 0.9876),
					"x2": Str("x2", "extra 1"),
					"i1": Int("i1", -89),
				},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true), Float("x1", 0.9876), Str("x2", "extra 1")},
					},
				},
			},
			args2: args{
				lvl:    LevelDebug,
				msg:    "2nd message",
				fields: FieldSet{Int("i2", 123), Dur("d2", 5*time.Second), Bool("b2", true)},
			},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true), Float("x1", 0.9876), Str("x2", "extra 1")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "2nd message",
						fields:  FieldSet{Int("i1", -89), Int("i2", 123), Dur("d2", 5*time.Second), Bool("b2", true), Float("x1", 0.9876), Str("x2", "extra 1")},
					},
				},
			},
		},
		{
			name: "multiple fields pre-recorded",
			fields: fields{
				recorded: []TestLogEvent{
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 1",
						fields:  FieldSet{Int("secret1", -7890), Str("secret2", "asdf")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 2",
						fields:  FieldSet{Int("secret3", 7891), Str("secret4", "qwerty")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 3",
						fields:  FieldSet{Int("i1", -89)},
					},
				},
				fieldsByKey: map[string]Field{
					"x1": Float("x1", 0.9876),
					"x2": Str("x2", "extra 1"),
					"i1": Int("i1", -89),
				},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 1",
						fields:  FieldSet{Int("secret1", -7890), Str("secret2", "asdf")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 2",
						fields:  FieldSet{Int("secret3", 7891), Str("secret4", "qwerty")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 3",
						fields:  FieldSet{Int("i1", -89)},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true), Float("x1", 0.9876), Str("x2", "extra 1")},
					},
				},
			},
			args2: args{
				lvl:    LevelDebug,
				msg:    "2nd message",
				fields: FieldSet{Int("i2", 123), Dur("d2", 5*time.Second), Bool("b2", true)},
			},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 1",
						fields:  FieldSet{Int("secret1", -7890), Str("secret2", "asdf")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 2",
						fields:  FieldSet{Int("secret3", 7891), Str("secret4", "qwerty")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "pre-recorded msg 3",
						fields:  FieldSet{Int("i1", -89)},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "a message",
						fields:  FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true), Float("x1", 0.9876), Str("x2", "extra 1")},
					},
					testLogEvent{
						level:   LevelDebug,
						message: "2nd message",
						fields:  FieldSet{Int("i1", -89), Int("i2", 123), Dur("d2", 5*time.Second), Bool("b2", true), Float("x1", 0.9876), Str("x2", "extra 1")},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testLogEventRecorder{
				recorded: tt.fields.recorded,
			}
			l := &testLoggerImpl{
				fieldsByKey: tt.fields.fieldsByKey,
				logFunc:     r.recordEvent,
			}

			l.log(tt.args.lvl, tt.args.msg, tt.args.fields...)

			normaliseTestLogEvents(tt.want.events, r.recorded)
			assert.ElementsMatch(t, tt.want.events, r.recorded)

			l.log(tt.args2.lvl, tt.args2.msg, tt.args2.fields...)

			normaliseTestLogEvents(tt.want2.events, r.recorded)
			assert.ElementsMatch(t, tt.want2.events, r.recorded)
		})
	}
}

func Test_testLoggerImpl_clone(t *testing.T) {
	type fields struct {
		recorded    []TestLogEvent
		fieldsByKey map[string]Field
	}
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	type want struct {
		events []TestLogEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		args2  args
		want   want
		want2  want
	}{
		{
			name: "add fields",
			fields: fields{
				recorded:    []TestLogEvent{},
				fieldsByKey: map[string]Field{},
			},
			args: args{lvl: LevelInfo, msg: "before clone", fields: FieldSet{}},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{},
					},
				},
			},
			args2: args{lvl: LevelInfo, msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "after clone",
						fields:  FieldSet{Int("i1", 123)},
					},
				},
			},
		},
		{
			name: "add fields pre-existing",
			fields: fields{
				recorded: []TestLogEvent{},
				fieldsByKey: map[string]Field{
					"x1": Str("x1", "secret 1"),
					"x2": Str("x2", "secret 2"),
				},
			},
			args: args{lvl: LevelInfo, msg: "before clone", fields: FieldSet{}},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
				},
			},
			args2: args{lvl: LevelInfo, msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "after clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2"), Int("i1", 123)},
					},
				},
			},
		},
		{
			name: "add fields pre-recorded",
			fields: fields{
				recorded: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 1",
						fields:  FieldSet{Str("rec1", "value1")},
					},
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 2",
						fields:  FieldSet{Str("rec2", "value2"), Int("rec3", 123)},
					},
				},
				fieldsByKey: map[string]Field{
					"x1": Str("x1", "secret 1"),
					"x2": Str("x2", "secret 2"),
				},
			},
			args: args{lvl: LevelInfo, msg: "before clone", fields: FieldSet{}},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 1",
						fields:  FieldSet{Str("rec1", "value1")},
					},
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 2",
						fields:  FieldSet{Str("rec2", "value2"), Int("rec3", 123)},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
				},
			},
			args2: args{lvl: LevelInfo, msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 1",
						fields:  FieldSet{Str("rec1", "value1")},
					},
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 2",
						fields:  FieldSet{Str("rec2", "value2"), Int("rec3", 123)},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "after clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2"), Int("i1", 123)},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testLogEventRecorder{
				recorded: tt.fields.recorded,
			}
			l := &testLoggerImpl{
				fieldsByKey: tt.fields.fieldsByKey,
				logFunc:     r.recordEvent,
			}

			lnew := l.clone()

			l.log(tt.args.lvl, tt.args.msg, tt.args.fields...)

			normaliseTestLogEvents(tt.want.events, r.recorded)
			assert.ElementsMatch(t, tt.want.events, r.recorded)

			lnew.log(tt.args2.lvl, tt.args2.msg, tt.args2.fields...)

			normaliseTestLogEvents(tt.want2.events, r.recorded)
			assert.ElementsMatch(t, tt.want2.events, r.recorded)
		})
	}
}

func Test_testLoggerImpl_with(t *testing.T) {
	type fields struct {
		recorded    []TestLogEvent
		fieldsByKey map[string]Field
	}
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	type want struct {
		events []TestLogEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		args2  args
		want   want
		want2  want
	}{
		{
			name: "add fields",
			fields: fields{
				recorded:    []TestLogEvent{},
				fieldsByKey: map[string]Field{},
			},
			args: args{lvl: LevelInfo, msg: "before clone", fields: FieldSet{}},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{},
					},
				},
			},
			args2: args{lvl: LevelInfo, msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "after clone",
						fields:  FieldSet{Int("i1", 123)},
					},
				},
			},
		},
		{
			name: "add fields pre-existing",
			fields: fields{
				recorded: []TestLogEvent{},
				fieldsByKey: map[string]Field{
					"x1": Str("x1", "secret 1"),
					"x2": Str("x2", "secret 2"),
				},
			},
			args: args{lvl: LevelInfo, msg: "before clone", fields: FieldSet{}},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
				},
			},
			args2: args{lvl: LevelInfo, msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "after clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2"), Int("i1", 123)},
					},
				},
			},
		},
		{
			name: "add fields pre-recorded",
			fields: fields{
				recorded: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 1",
						fields:  FieldSet{Str("rec1", "value1")},
					},
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 2",
						fields:  FieldSet{Str("rec2", "value2"), Int("rec3", 123)},
					},
				},
				fieldsByKey: map[string]Field{
					"x1": Str("x1", "secret 1"),
					"x2": Str("x2", "secret 2"),
				},
			},
			args: args{lvl: LevelInfo, msg: "before clone", fields: FieldSet{}},
			want: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 1",
						fields:  FieldSet{Str("rec1", "value1")},
					},
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 2",
						fields:  FieldSet{Str("rec2", "value2"), Int("rec3", 123)},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
				},
			},
			args2: args{lvl: LevelInfo, msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want2: want{
				events: []TestLogEvent{
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 1",
						fields:  FieldSet{Str("rec1", "value1")},
					},
					testLogEvent{
						level:   LevelError,
						message: "pre-recorded 2",
						fields:  FieldSet{Str("rec2", "value2"), Int("rec3", 123)},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "before clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2")},
					},
					testLogEvent{
						level:   LevelInfo,
						message: "after clone",
						fields:  FieldSet{Str("x1", "secret 1"), Str("x2", "secret 2"), Int("i1", 123)},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &testLogEventRecorder{
				recorded: tt.fields.recorded,
			}
			l := &testLoggerImpl{
				fieldsByKey: tt.fields.fieldsByKey,
				logFunc:     r.recordEvent,
			}

			lnew := l.with(tt.args2.fields...)

			l.log(tt.args.lvl, tt.args.msg, tt.args.fields...)

			normaliseTestLogEvents(tt.want.events, r.recorded)
			assert.ElementsMatch(t, tt.want.events, r.recorded)

			lnew.log(tt.args2.lvl, tt.args2.msg)

			normaliseTestLogEvents(tt.want2.events, r.recorded)
			assert.ElementsMatch(t, tt.want2.events, r.recorded)
		})
	}
}

func Test_testLoggerImpl_sync(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "default",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &testLoggerImpl{}
			err := l.sync()
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_testLogger_With(t *testing.T) {
	type fields struct {
		baseLogger *baseLogger
		recorder   *testLogEventRecorder
	}
	type args struct {
		fields FieldSet
	}
	type want struct {
		fields FieldSet
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion func(*testing.T, fields, want, Logger)
	}{
		{
			name: "default",
			fields: fields{
				baseLogger: &baseLogger{
					impl: &testLoggerImpl{},
				},
				recorder: &testLogEventRecorder{},
			},
			args: args{
				fields: FieldSet{Int("i1", 123)},
			},
			want: want{
				fields: FieldSet{Int("i1", 123)},
			},
			assertion: func(t *testing.T, fields fields, want want, got Logger) {
				require.IsType(t, &testLogger{}, got)
				l := got.(*testLogger)
				assert.NotEqual(t, fields.baseLogger, l.baseLogger)
				assert.Equal(t, fields.recorder, l.recorder)
				require.IsType(t, &testLoggerImpl{}, l.impl)
				impl := l.impl.(*testLoggerImpl)
				assert.Equal(t, want.fields.ToMap(), impl.fieldsByKey)
				assert.Len(t, fields.baseLogger.impl.(*testLoggerImpl).fieldsByKey, 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &testLogger{
				baseLogger: tt.fields.baseLogger,
				recorder:   tt.fields.recorder,
			}
			tt.assertion(t, tt.fields, tt.want, l.With(tt.args.fields...))
		})
	}
}

func Test_testLogEvent_ContainsFields(t *testing.T) {
	type fields struct {
		level   Level
		message string
		fields  FieldSet
	}
	type args struct {
		subset []Field
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testLogEvent{
				level:   tt.fields.level,
				message: tt.fields.message,
				fields:  tt.fields.fields,
			}
			assert.Equal(t, tt.want, e.ContainsFields(tt.args.subset...))
		})
	}
}

func TestTestLogEventSet_FindMatchingEvents(t *testing.T) {
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	tests := []struct {
		name string
		es   TestLogEventSet
		args args
		want TestLogEventSet
	}{
		{
			name: "empty set, no match",
			es:   TestLogEventSet{},
			args: args{
				lvl:    LevelDebug,
				msg:    "not found",
				fields: FieldSet{},
			},
			want: TestLogEventSet{},
		},
		{
			name: "no match",
			es: TestLogEventSet{
				testLogEvent{
					level:   LevelError,
					message: "msg error",
					fields: FieldSet{
						Int("i1", 123),
						Float("f1", 1.23),
					},
				},
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
			},
			args: args{
				lvl:    LevelDebug,
				msg:    "not found",
				fields: FieldSet{},
			},
			want: TestLogEventSet{},
		},
		{
			name: "no match 2",
			es: TestLogEventSet{
				testLogEvent{
					level:   LevelError,
					message: "msg error",
					fields: FieldSet{
						Int("i1", 123),
						Float("f1", 1.23),
					},
				},
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
			},
			args: args{
				lvl: LevelInfo,
				msg: "msg info",
				fields: FieldSet{
					Int("i1", 789),
				},
			},
			want: TestLogEventSet{},
		},
		{
			name: "no match 3",
			es: TestLogEventSet{
				testLogEvent{
					level:   LevelError,
					message: "msg error",
					fields: FieldSet{
						Int("i1", 123),
						Float("f1", 1.23),
					},
				},
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
			},
			args: args{
				lvl: LevelInfo,
				msg: "no way!",
				fields: FieldSet{
					Int("i1", 789),
				},
			},
			want: TestLogEventSet{},
		},
		{
			name: "full match",
			es: TestLogEventSet{
				testLogEvent{
					level:   LevelError,
					message: "msg error",
					fields: FieldSet{
						Int("i1", 123),
						Float("f1", 1.23),
					},
				},
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
				testLogEvent{
					level:   LevelDebug,
					message: "msg debug",
					fields: FieldSet{
						Int("i1", 789),
						Float("f1", -1.23),
					},
				},
			},
			args: args{
				lvl: LevelInfo,
				msg: "msg info",
				fields: FieldSet{
					Float("f1", 4.56),
					Int("i1", 456),
				},
			},
			want: TestLogEventSet{
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
			},
		},
		{
			name: "partial match",
			es: TestLogEventSet{
				testLogEvent{
					level:   LevelError,
					message: "msg error",
					fields: FieldSet{
						Int("i1", 123),
						Float("f1", 1.23),
					},
				},
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
				testLogEvent{
					level:   LevelDebug,
					message: "msg debug",
					fields: FieldSet{
						Int("i1", 789),
						Float("f1", -1.23),
						Str("x1", "a needle"),
						Float("f3", 0.33),
					},
				},
			},
			args: args{
				lvl: LevelDebug,
				msg: "msg debug",
				fields: FieldSet{
					Str("x1", "a needle"),
				},
			},
			want: TestLogEventSet{
				testLogEvent{
					level:   LevelDebug,
					message: "msg debug",
					fields: FieldSet{
						Int("i1", 789),
						Float("f1", -1.23),
						Str("x1", "a needle"),
						Float("f3", 0.33),
					},
				},
			},
		},
		{
			name: "partial match 2",
			es: TestLogEventSet{
				testLogEvent{
					level:   LevelError,
					message: "msg error",
					fields: FieldSet{
						Int("i1", 123),
						Float("f1", 1.23),
					},
				},
				testLogEvent{
					level:   LevelInfo,
					message: "msg info",
					fields: FieldSet{
						Int("i1", 456),
						Float("f1", 4.56),
					},
				},
				testLogEvent{
					level:   LevelDebug,
					message: "msg debug",
					fields: FieldSet{
						Int("i1", 789),
						Float("f1", -1.23),
						Str("x1", "a needle"),
						Float("f3", 0.33),
					},
				},
			},
			args: args{
				lvl: LevelDebug,
				msg: "msg debug",
				fields: FieldSet{
					Str("x1", "a needle"),
					Int("i1", 789),
				},
			},
			want: TestLogEventSet{
				testLogEvent{
					level:   LevelDebug,
					message: "msg debug",
					fields: FieldSet{
						Int("i1", 789),
						Float("f1", -1.23),
						Str("x1", "a needle"),
						Float("f3", 0.33),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.es.FindMatchingEvents(tt.args.lvl, tt.args.msg, tt.args.fields...))
		})
	}
}
