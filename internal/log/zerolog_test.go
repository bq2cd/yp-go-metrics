package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZeroLogger(t *testing.T) {
	tests := []struct {
		name      string
		writer    io.Writer
		assertion func(*testing.T, io.Writer, Logger, error)
	}{
		{
			name:   "bytes buffer",
			writer: &bytes.Buffer{},
			assertion: func(t *testing.T, want io.Writer, got Logger, err error) {
				assert.NoError(t, err)
				assert.IsType(t, &baseLogger{}, got)
				assert.IsType(t, &zeroLogger{}, got.(*baseLogger).impl)
				assert.IsType(t, zerolog.Logger{}, got.(*baseLogger).impl.(*zeroLogger).logger)
				zlog := got.(*baseLogger).impl.(*zeroLogger).logger
				zlog.Log().Msg("test")
				assert.JSONEq(t, `{"msg":"test"}`, want.(*bytes.Buffer).String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewZeroLogger(tt.writer)
			tt.assertion(t, tt.writer, got, err)
		})
	}
}

func Test_convertLevelToZero(t *testing.T) {
	type args struct {
		lvl Level
	}
	tests := []struct {
		name string
		args args
		want zerolog.Level
	}{
		{
			name: "fatal",
			args: args{lvl: LevelFatal},
			want: zerolog.FatalLevel,
		},
		{
			name: "error",
			args: args{lvl: LevelError},
			want: zerolog.ErrorLevel,
		},
		{
			name: "info",
			args: args{lvl: LevelInfo},
			want: zerolog.InfoLevel,
		},
		{
			name: "debug",
			args: args{lvl: LevelDebug},
			want: zerolog.DebugLevel,
		},
		{
			name: "unknown",
			args: args{lvl: Level(-99)},
			want: zerolog.TraceLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, convertLevelToZero(tt.args.lvl))
		})
	}
}

func Test_convertFieldToZeroEvent(t *testing.T) {
	type args struct {
		field Field
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "bool",
			args: args{field: Bool("key", true)},
			want: map[string]any{"key": true},
		},
		{
			name: "int",
			args: args{field: Int("key", -123)},
			want: map[string]any{"key": -123},
		},
		{
			name: "float",
			args: args{field: Float("key", -0.05)},
			want: map[string]any{"key": -0.05},
		},
		{
			name: "string",
			args: args{field: Str("key", "value")},
			want: map[string]any{"key": "value"},
		},
		{
			name: "error",
			args: args{field: Err("key", errors.New("an error"))},
			want: map[string]any{"error": "an error"},
		},
		{
			name: "duration",
			args: args{field: Dur("key", 3*time.Second)},
			want: map[string]any{"key": 3000},
		},
		{
			name: "time",
			args: args{field: Time("key", time.Date(2025, time.October, 1, 1, 2, 3, 0, time.UTC))},
			want: map[string]any{"key": 1759280523000000.0},
		},
		{
			name: "any",
			args: args{field: Any("key", map[string][]string{"innerK": {"innerV1", "innerV2"}})},
			want: map[string]any{"key": map[string][]string{"innerK": {"innerV1", "innerV2"}}},
		},
		{
			name: "unknown",
			args: args{field: Field{Key: "key", Type: FieldType(0), Value: struct {
				val1 string
				val2 int
				val3 bool
			}{"s", 3, true}}},
			want: map[string]any{"key": struct {
				val1 string
				val2 int
				val3 bool
			}{"s", 3, true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := zerolog.New(&buf)
			evt := convertFieldToZeroEvent(l.Log(), tt.args.field)
			evt.Send()
			want, err := json.Marshal(tt.want)
			assert.NoError(t, err)
			assert.JSONEq(t, string(want), buf.String())
		})
	}
}

func Test_zeroLogger_addFields(t *testing.T) {
	type fields struct {
		fieldsByKey map[string]Field
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
			l := &zeroLogger{
				logger:      zerolog.New(io.Discard),
				fieldsByKey: tt.fields.fieldsByKey,
			}
			l.addFields(tt.args.fields...)
			assert.Equal(t, tt.want.fieldsByKey, l.fieldsByKey)
		})
	}
}

func Test_zeroLogger_logFatal_exit(t *testing.T) {
	if os.Getenv("TEST_ZERO_LOGGER_LOG_FATAL") != "1" {
		return
	}

	l := &zeroLogger{
		logger:      zerolog.New(os.Stderr).Level(zerolog.FatalLevel),
		fieldsByKey: map[string]Field{},
	}

	l.log(LevelFatal, "oops!", Bool("crashed", true))
}

func Test_zeroLogger_logFatal(t *testing.T) {
	var buf bytes.Buffer

	cmd := exec.Command(os.Args[0], "-test.run=Test_zeroLogger_logFatal_exit")
	cmd.Env = append(os.Environ(), "TEST_ZERO_LOGGER_LOG_FATAL=1")
	cmd.Stdout = io.Discard
	cmd.Stderr = &buf

	err := cmd.Run()

	require.IsType(t, &exec.ExitError{}, err)
	status := err.(*exec.ExitError)

	assert.Equal(t, 1, status.ExitCode())

	var got map[string]any
	err = json.Unmarshal(buf.Bytes(), &got)
	assert.NoError(t, err)
	delete(got, "caller")
	delete(got, "ts")
	assert.Equal(t, map[string]any{"level": "fatal", "msg": "oops!", "crashed": true}, got)
}

func Test_zeroLogger_log(t *testing.T) {
	type fields struct {
		level       zerolog.Level
		fieldsByKey map[string]Field
	}
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	type want struct {
		skipped  bool
		logEntry map[string]any
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
			name: "debug level logs debug msg",
			fields: fields{
				level:       zerolog.DebugLevel,
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelDebug,
				msg:    "debug me",
				fields: FieldSet{},
			},
			want: want{
				logEntry: map[string]any{
					"level": "debug",
					"msg":   "debug me",
				},
			},
		},
		{
			name: "info level skips debug msg",
			fields: fields{
				level:       zerolog.InfoLevel,
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelDebug,
				msg:    "debug me",
				fields: FieldSet{},
			},
			want: want{
				skipped: true,
			},
		},
		{
			name: "info level logs error msg",
			fields: fields{
				level:       zerolog.InfoLevel,
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelError,
				msg:    "not again",
				fields: FieldSet{Err("err", errors.New("an error"))},
			},
			want: want{
				logEntry: map[string]any{
					"level": "error",
					"msg":   "not again",
					"error": "an error",
				},
			},
		},
		{
			name: "multiple fields",
			fields: fields{
				level:       zerolog.InfoLevel,
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				logEntry: map[string]any{
					"level": "info",
					"msg":   "a message",
					"i1":    float64(123),
					"d1":    float64(5000),
					"b1":    true,
				},
			},
		},
		{
			name: "multiple fields pre-existing",
			fields: fields{
				level: zerolog.InfoLevel,
				fieldsByKey: map[string]Field{
					"i1": Int("i1", -123),
				},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				logEntry: map[string]any{
					"level": "info",
					"msg":   "a message",
					"i1":    float64(123),
					"d1":    float64(5000),
					"b1":    true,
				},
			},
		},
		{
			name: "duplicate fields pre-existing",
			fields: fields{
				level: zerolog.InfoLevel,
				fieldsByKey: map[string]Field{
					"f1": Float("i1", -1.23),
				},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Float("f1", 1.23), Bool("b1", true), Float("f1", -4.56)},
			},
			want: want{
				logEntry: map[string]any{
					"level": "info",
					"msg":   "a message",
					"i1":    float64(123),
					"f1":    float64(-4.56),
					"b1":    true,
				},
			},
		},
		{
			name: "multiple fields 2-phase",
			fields: fields{
				level:       zerolog.InfoLevel,
				fieldsByKey: map[string]Field{},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				logEntry: map[string]any{
					"level": "info",
					"msg":   "a message",
					"i1":    float64(123),
					"d1":    float64(5000),
					"b1":    true,
				},
			},
			args2: args{
				lvl:    LevelInfo,
				msg:    "a message 2",
				fields: FieldSet{Int("i2", 123456), Dur("d2", 15*time.Second), Bool("b2", false)},
			},
			want2: want{
				logEntry: map[string]any{
					"level": "info",
					"msg":   "a message 2",
					"i2":    float64(123456),
					"d2":    float64(15000),
					"b2":    false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := &zeroLogger{
				logger:      zerolog.New(&buf).Level(tt.fields.level),
				fieldsByKey: tt.fields.fieldsByKey,
			}

			l.log(tt.args.lvl, tt.args.msg, tt.args.fields...)

			if tt.want.skipped {
				assert.Len(t, buf.Bytes(), 0)
				return
			}

			got := map[string]any{}
			err := json.Unmarshal(buf.Bytes(), &got)
			assert.NoError(t, err)
			delete(got, "caller")
			delete(got, "ts")
			assert.Equal(t, tt.want.logEntry, got)

			if len(tt.args2.fields) == 0 && tt.args2.msg == "" {
				return
			}
			buf.Reset()

			l.log(tt.args2.lvl, tt.args2.msg, tt.args2.fields...)

			got = map[string]any{}
			err = json.Unmarshal(buf.Bytes(), &got)
			assert.NoError(t, err)
			delete(got, "caller")
			delete(got, "ts")
			assert.Equal(t, tt.want2.logEntry, got)
		})
	}
}

func Test_zeroLogger_clone(t *testing.T) {
	type fields struct {
		fieldsByKey map[string]Field
	}
	type args struct {
		msg    string
		fields FieldSet
	}
	type want struct {
		logEntry map[string]any
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		argsClone args
		want      want
		wantClone want
	}{
		{
			name: "add fields",
			fields: fields{
				fieldsByKey: map[string]Field{},
			},
			args:      args{msg: "before clone", fields: FieldSet{}},
			argsClone: args{msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want: want{logEntry: map[string]any{
				"msg": "before clone",
			}},
			wantClone: want{logEntry: map[string]any{
				"msg": "after clone",
				"i1":  float64(123),
			}},
		},
		{
			name: "add fields pre-existing",
			fields: fields{
				fieldsByKey: map[string]Field{
					"b1": Bool("b1", true),
				},
			},
			args:      args{msg: "before clone", fields: FieldSet{}},
			argsClone: args{msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want: want{logEntry: map[string]any{
				"msg": "before clone",
				"b1":  true,
			}},
			wantClone: want{logEntry: map[string]any{
				"msg": "after clone",
				"b1":  true,
				"i1":  float64(123),
			}},
		},
		{
			name: "add duplicate fields pre-existing",
			fields: fields{
				fieldsByKey: map[string]Field{
					"b1": Bool("b1", true),
					"f1": Float("f1", -4.56),
				},
			},
			args:      args{msg: "before clone", fields: FieldSet{Float("f1", 1.23)}},
			argsClone: args{msg: "after clone", fields: FieldSet{Int("i1", 123), Float("f1", -0.123)}},
			want: want{logEntry: map[string]any{
				"msg": "before clone",
				"b1":  true,
				"f1":  float64(1.23),
			}},
			wantClone: want{logEntry: map[string]any{
				"msg": "after clone",
				"b1":  true,
				"i1":  float64(123),
				"f1":  float64(-0.123),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := &zeroLogger{
				logger:      zerolog.New(&buf).Level(zerolog.InfoLevel),
				fieldsByKey: tt.fields.fieldsByKey,
			}

			lnew := l.clone()
			l.log(LevelInfo, tt.args.msg, tt.args.fields...)
			origMsg := make([]byte, buf.Len())
			copy(origMsg, buf.Bytes())
			buf.Reset()
			lnew.log(LevelInfo, tt.argsClone.msg, tt.argsClone.fields...)
			newMsg := make([]byte, buf.Len())
			copy(newMsg, buf.Bytes())

			var origEntry map[string]any
			err := json.Unmarshal(origMsg, &origEntry)
			assert.NoError(t, err)
			delete(origEntry, "caller")
			delete(origEntry, "ts")
			delete(origEntry, "level")
			assert.Equal(t, tt.want.logEntry, origEntry)

			var newEntry map[string]any
			err = json.Unmarshal(newMsg, &newEntry)
			assert.NoError(t, err)
			delete(newEntry, "caller")
			delete(newEntry, "ts")
			delete(newEntry, "level")
			assert.Equal(t, tt.wantClone.logEntry, newEntry)
		})
	}
}

func Test_zeroLogger_with(t *testing.T) {
	type fields struct {
		fieldsByKey map[string]Field
	}
	type args struct {
		msg    string
		fields FieldSet
	}
	type want struct {
		logEntry map[string]any
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		argsWith args
		want     want
		wantWith want
	}{
		{
			name: "add fields",
			fields: fields{
				fieldsByKey: map[string]Field{},
			},
			args:     args{msg: "before with", fields: FieldSet{}},
			argsWith: args{msg: "after with", fields: FieldSet{Int("i1", 123)}},
			want: want{logEntry: map[string]any{
				"msg": "before with",
			}},
			wantWith: want{logEntry: map[string]any{
				"msg": "after with",
				"i1":  float64(123),
			}},
		},
		{
			name: "add fields pre-existing",
			fields: fields{
				fieldsByKey: map[string]Field{
					"b1": Bool("b1", true),
				},
			},
			args:     args{msg: "before with", fields: FieldSet{}},
			argsWith: args{msg: "after with", fields: FieldSet{Int("i1", 123)}},
			want: want{logEntry: map[string]any{
				"msg": "before with",
				"b1":  true,
			}},
			wantWith: want{logEntry: map[string]any{
				"msg": "after with",
				"b1":  true,
				"i1":  float64(123),
			}},
		},
		{
			name: "add duplicate fields pre-existing",
			fields: fields{
				fieldsByKey: map[string]Field{
					"b1": Bool("b1", true),
					"f1": Float("f1", -4.56),
				},
			},
			args:     args{msg: "before with", fields: FieldSet{Float("f1", 1.23)}},
			argsWith: args{msg: "after with", fields: FieldSet{Int("i1", 123), Float("f1", -0.123)}},
			want: want{logEntry: map[string]any{
				"msg": "before with",
				"b1":  true,
				"f1":  float64(1.23),
			}},
			wantWith: want{logEntry: map[string]any{
				"msg": "after with",
				"b1":  true,
				"i1":  float64(123),
				"f1":  float64(-0.123),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := &zeroLogger{
				logger:      zerolog.New(&buf).Level(zerolog.InfoLevel),
				fieldsByKey: tt.fields.fieldsByKey,
			}

			lnew := l.with(tt.argsWith.fields...)
			l.log(LevelInfo, tt.args.msg, tt.args.fields...)
			origMsg := make([]byte, buf.Len())
			copy(origMsg, buf.Bytes())
			buf.Reset()
			lnew.log(LevelInfo, tt.argsWith.msg)
			newMsg := make([]byte, buf.Len())
			copy(newMsg, buf.Bytes())

			var origEntry map[string]any
			err := json.Unmarshal(origMsg, &origEntry)
			assert.NoError(t, err)
			delete(origEntry, "caller")
			delete(origEntry, "ts")
			delete(origEntry, "level")
			assert.Equal(t, tt.want.logEntry, origEntry)

			var newEntry map[string]any
			err = json.Unmarshal(newMsg, &newEntry)
			assert.NoError(t, err)
			delete(newEntry, "caller")
			delete(newEntry, "ts")
			delete(newEntry, "level")
			assert.Equal(t, tt.wantWith.logEntry, newEntry)
		})
	}
}

func Test_zeroLogger_sync(t *testing.T) {
	type fields struct {
		logger      zerolog.Logger
		fieldsByKey map[string]Field
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "default",
			fields: fields{
				logger:      zerolog.New(io.Discard),
				fieldsByKey: map[string]Field{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &zeroLogger{
				logger:      tt.fields.logger,
				fieldsByKey: tt.fields.fieldsByKey,
			}
			assert.NoError(t, l.sync())
		})
	}
}
