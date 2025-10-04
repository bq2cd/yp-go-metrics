package log

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewZapLogger(t *testing.T) {
	type args struct {
		cfg zap.Config
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, Logger, error)
	}{
		{
			name: "empty config",
			args: args{
				cfg: zap.Config{},
			},
			assertion: func(t *testing.T, got Logger, err error) {
				assert.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name: "development config",
			args: args{
				cfg: zap.NewDevelopmentConfig(),
			},
			assertion: func(t *testing.T, got Logger, err error) {
				assert.NoError(t, err)
				l := got.(*baseLogger).impl.(*zapLogger).logger
				assert.Equal(t, zap.DebugLevel, l.Level())
			},
		},
		{
			name: "production config",
			args: args{
				cfg: zap.NewProductionConfig(),
			},
			assertion: func(t *testing.T, got Logger, err error) {
				assert.NoError(t, err)
				l := got.(*baseLogger).impl.(*zapLogger).logger
				assert.Equal(t, zap.InfoLevel, l.Level())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewZapLogger(tt.args.cfg)
			tt.assertion(t, got, err)
		})
	}
}

func Test_convertLevelToZap(t *testing.T) {
	type args struct {
		lvl Level
	}
	tests := []struct {
		name string
		args args
		want zapcore.Level
	}{
		{
			name: "fatal",
			args: args{lvl: LevelFatal},
			want: zapcore.FatalLevel,
		},
		{
			name: "error",
			args: args{lvl: LevelError},
			want: zapcore.ErrorLevel,
		},
		{
			name: "info",
			args: args{lvl: LevelInfo},
			want: zapcore.InfoLevel,
		},
		{
			name: "debug",
			args: args{lvl: LevelDebug},
			want: zapcore.DebugLevel,
		},
		{
			name: "unknown",
			args: args{lvl: Level(-99)},
			want: zapcore.DebugLevel - 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, convertLevelToZap(tt.args.lvl))
		})
	}
}

func Test_convertFieldToZap(t *testing.T) {
	type args struct {
		field Field
	}
	tests := []struct {
		name string
		args args
		want zap.Field
	}{
		{
			name: "bool",
			args: args{field: Bool("key", true)},
			want: zap.Bool("key", true),
		},
		{
			name: "int",
			args: args{field: Int("key", -123)},
			want: zap.Int("key", -123),
		},
		{
			name: "float",
			args: args{field: Float("key", -0.05)},
			want: zap.Float64("key", -0.05),
		},
		{
			name: "string",
			args: args{field: Str("key", "value")},
			want: zap.String("key", "value"),
		},
		{
			name: "error",
			args: args{field: Err("key", errors.New("an error"))},
			want: zap.NamedError("key", errors.New("an error")),
		},
		{
			name: "duration",
			args: args{field: Dur("key", 3*time.Second)},
			want: zap.Duration("key", 3*time.Second),
		},
		{
			name: "time",
			args: args{field: Time("key", time.Date(2025, time.October, 1, 1, 2, 3, 0, time.UTC))},
			want: zap.Time("key", time.Date(2025, time.October, 1, 1, 2, 3, 0, time.UTC)),
		},
		{
			name: "any",
			args: args{field: Any("key", map[string][]string{"innerK": {"innerV1", "innerV2"}})},
			want: zap.Any("key", map[string][]string{"innerK": {"innerV1", "innerV2"}}),
		},
		{
			name: "unknown",
			args: args{field: Field{Key: "key", Type: FieldType(0), Value: struct {
				val1 string
				val2 int
				val3 bool
			}{"s", 3, true}}},
			want: zap.Any("key", struct {
				val1 string
				val2 int
				val3 bool
			}{"s", 3, true}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, convertFieldToZap(tt.args.field))
		})
	}
}

func Test_zapLogger_log(t *testing.T) {
	type fields struct {
		level   zapcore.Level
		options []zap.Option
	}
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	type want struct {
		skipped  bool
		lvl      zapcore.Level
		msg      string
		fieldMap map[string]any
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion func(*testing.T, *observer.ObservedLogs)
	}{
		{
			name: "debug level logs debug msg",
			fields: fields{
				level: zap.DebugLevel,
			},
			args: args{
				lvl:    LevelDebug,
				msg:    "a message",
				fields: FieldSet{},
			},
			want: want{
				lvl:      zapcore.DebugLevel,
				msg:      "a message",
				fieldMap: map[string]any{},
			},
		},
		{
			name: "info level skips debug msg",
			fields: fields{
				level: zap.InfoLevel,
			},
			args: args{
				lvl:    LevelDebug,
				msg:    "a message",
				fields: FieldSet{},
			},
			want: want{
				skipped: true,
			},
		},
		{
			name: "info level logs error msg",
			fields: fields{
				level: zap.InfoLevel,
			},
			args: args{
				lvl:    LevelError,
				msg:    "a message",
				fields: FieldSet{Err("err", errors.New("an error"))},
			},
			want: want{
				lvl: zapcore.ErrorLevel,
				msg: "a message",
				fieldMap: map[string]any{
					"err": "an error",
				},
			},
		},
		{
			name: "multiple fields",
			fields: fields{
				level: zap.InfoLevel,
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				lvl: zapcore.InfoLevel,
				msg: "a message",
				fieldMap: map[string]any{
					"i1": int64(123),
					"d1": 5 * time.Second,
					"b1": true,
				},
			},
		},
		{
			name: "multiple fields pre-existing",
			fields: fields{
				level:   zap.InfoLevel,
				options: []zap.Option{zap.Fields(zap.Float64("f1", 1.23))},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Dur("d1", 5*time.Second), Bool("b1", true)},
			},
			want: want{
				lvl: zapcore.InfoLevel,
				msg: "a message",
				fieldMap: map[string]any{
					"f1": float64(1.23),
					"i1": int64(123),
					"d1": 5 * time.Second,
					"b1": true,
				},
			},
		},
		{
			name: "duplicate fields",
			fields: fields{
				level: zap.InfoLevel,
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Float("f1", 1.23), Bool("b1", true), Float("f1", -4.56)},
			},
			want: want{
				lvl: zapcore.InfoLevel,
				msg: "a message",
				fieldMap: map[string]any{
					"i1": int64(123),
					"b1": true,
					"f1": float64(-4.56),
				},
			},
		},
		{
			name: "duplicate fields pre-existing",
			fields: fields{
				level:   zap.InfoLevel,
				options: []zap.Option{zap.Fields(zap.Float64("f1", -1.23))},
			},
			args: args{
				lvl:    LevelInfo,
				msg:    "a message",
				fields: FieldSet{Int("i1", 123), Float("f1", 1.23), Bool("b1", true), Float("f1", -4.56)},
			},
			want: want{
				lvl: zapcore.InfoLevel,
				msg: "a message",
				fieldMap: map[string]any{
					"i1": int64(123),
					"b1": true,
					"f1": float64(-4.56),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(tt.fields.level)
			l := &zapLogger{
				logger: zap.New(core, tt.fields.options...),
			}

			l.log(tt.args.lvl, tt.args.msg, tt.args.fields...)

			if tt.want.skipped {
				assert.Equal(t, 0, logs.Len())
				return
			}
			assert.Equal(t, 1, logs.Len())
			got := logs.AllUntimed()[0]
			assert.Equal(t, tt.want.lvl, got.Level)
			assert.Equal(t, tt.want.msg, got.Message)
			assert.Equal(t, tt.want.fieldMap, got.ContextMap())
		})
	}
}

func Test_zapLogger_clone(t *testing.T) {
	type fields struct {
		options []zap.Option
	}
	type args struct {
		msg    string
		fields FieldSet
	}
	type want struct {
		msg      string
		fieldMap map[string]any
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
			name:      "add fields",
			fields:    fields{},
			args:      args{msg: "before clone", fields: FieldSet{}},
			argsClone: args{msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want:      want{msg: "before clone", fieldMap: map[string]any{}},
			wantClone: want{msg: "after clone", fieldMap: map[string]any{"i1": int64(123)}},
		},
		{
			name:      "add fields pre-existing",
			fields:    fields{options: []zap.Option{zap.Fields(zap.Bool("b1", true))}},
			args:      args{msg: "before clone", fields: FieldSet{}},
			argsClone: args{msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want:      want{msg: "before clone", fieldMap: map[string]any{"b1": true}},
			wantClone: want{msg: "after clone", fieldMap: map[string]any{"b1": true, "i1": int64(123)}},
		},
		{
			name:      "add duplicate fields pre-existing",
			fields:    fields{options: []zap.Option{zap.Fields(zap.Bool("b1", true), zap.Int("i1", 123))}},
			args:      args{msg: "before clone", fields: FieldSet{}},
			argsClone: args{msg: "after clone", fields: FieldSet{Int("i1", -123)}},
			want:      want{msg: "before clone", fieldMap: map[string]any{"b1": true, "i1": int64(123)}},
			wantClone: want{msg: "after clone", fieldMap: map[string]any{"b1": true, "i1": int64(-123)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			l := &zapLogger{
				logger: zap.New(core, tt.fields.options...),
			}

			lnew := l.clone()
			l.log(LevelInfo, tt.args.msg, tt.args.fields...)
			lnew.log(LevelInfo, tt.argsClone.msg, tt.argsClone.fields...)

			assert.Equal(t, 2, logs.Len())
			origLogs := logs.FilterMessage(tt.args.msg)
			assert.Equal(t, 1, origLogs.Len())
			assert.Equal(t, tt.want.msg, origLogs.All()[0].Message)
			assert.Equal(t, tt.want.fieldMap, origLogs.All()[0].ContextMap())
			newLogs := logs.FilterMessage(tt.argsClone.msg)
			assert.Equal(t, 1, newLogs.Len())
			assert.Equal(t, tt.wantClone.msg, newLogs.All()[0].Message)
			assert.Equal(t, tt.wantClone.fieldMap, newLogs.All()[0].ContextMap())
		})
	}
}

func Test_zapLogger_with(t *testing.T) {
	type fields struct {
		options []zap.Option
	}
	type args struct {
		msg    string
		fields FieldSet
	}
	type want struct {
		msg      string
		fieldMap map[string]any
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
			name:     "add fields",
			fields:   fields{},
			args:     args{msg: "before with", fields: FieldSet{}},
			argsWith: args{msg: "after with", fields: FieldSet{Int("i1", 123)}},
			want:     want{msg: "before with", fieldMap: map[string]any{}},
			wantWith: want{msg: "after with", fieldMap: map[string]any{"i1": int64(123)}},
		},
		{
			name:     "add fields pre-existing",
			fields:   fields{options: []zap.Option{zap.Fields(zap.Bool("b1", true))}},
			args:     args{msg: "before clone", fields: FieldSet{}},
			argsWith: args{msg: "after clone", fields: FieldSet{Int("i1", 123)}},
			want:     want{msg: "before clone", fieldMap: map[string]any{"b1": true}},
			wantWith: want{msg: "after clone", fieldMap: map[string]any{"b1": true, "i1": int64(123)}},
		},
		{
			name:     "add duplicate fields pre-existing",
			fields:   fields{options: []zap.Option{zap.Fields(zap.Bool("b1", true), zap.Int("i1", 123))}},
			args:     args{msg: "before clone", fields: FieldSet{}},
			argsWith: args{msg: "after clone", fields: FieldSet{Int("i1", -123)}},
			want:     want{msg: "before clone", fieldMap: map[string]any{"b1": true, "i1": int64(123)}},
			wantWith: want{msg: "after clone", fieldMap: map[string]any{"b1": true, "i1": int64(-123)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			l := &zapLogger{
				logger: zap.New(core, tt.fields.options...),
			}

			lnew := l.with(tt.argsWith.fields...)
			l.log(LevelInfo, tt.args.msg, tt.args.fields...)
			lnew.log(LevelInfo, tt.argsWith.msg)

			assert.Equal(t, 2, logs.Len())
			origLogs := logs.FilterMessage(tt.args.msg)
			assert.Equal(t, 1, origLogs.Len())
			assert.Equal(t, tt.want.msg, origLogs.All()[0].Message)
			assert.Equal(t, tt.want.fieldMap, origLogs.All()[0].ContextMap())
			newLogs := logs.FilterMessage(tt.argsWith.msg)
			assert.Equal(t, 1, newLogs.Len())
			assert.Equal(t, tt.wantWith.msg, newLogs.All()[0].Message)
			assert.Equal(t, tt.wantWith.fieldMap, newLogs.All()[0].ContextMap())
		})
	}
}

type mockCore struct {
	mock.Mock
	syncError bool
}

func (m *mockCore) Enabled(lvl zapcore.Level) bool {
	m.Called(lvl)
	return true
}
func (m *mockCore) With(fields []zapcore.Field) zapcore.Core {
	m.Called(fields)
	return zapcore.NewNopCore()
}
func (m *mockCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	m.Called(e, ce)
	return ce
}
func (m *mockCore) Write(e zapcore.Entry, fields []zapcore.Field) error {
	m.Called(m, fields)
	return nil
}
func (m *mockCore) Sync() error {
	m.Called()
	if m.syncError {
		return errors.New("sync error")
	}
	return nil
}

func Test_zapLogger_sync(t *testing.T) {
	tests := []struct {
		name    string
		core    *mockCore
		wantErr error
	}{
		{
			name: "default",
			core: &mockCore{},
		},
		{
			name: "core error",
			core: &mockCore{syncError: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &zapLogger{
				logger: zap.New(tt.core),
			}
			tt.core.On("Sync").Return(tt.wantErr)

			_ = l.sync()

			tt.core.AssertExpectations(t)
		})
	}
}
