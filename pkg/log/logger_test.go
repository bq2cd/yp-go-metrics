package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLoggerImpl struct {
	mock.Mock

	fields FieldSet
}

func (m *mockLoggerImpl) clone() loggerImpl {
	m.Called()
	return &mockLoggerImpl{}
}

func (m *mockLoggerImpl) log(lvl Level, msg string, fields ...Field) {
	args := make([]any, len(fields)+2)
	args[0] = lvl
	args[1] = msg
	for i, f := range fields {
		args[i+2] = f
	}
	m.Called(args)
	m.fields = append(m.fields, fields...)
}

func (m *mockLoggerImpl) with(fields ...Field) loggerImpl {
	args := make([]any, len(fields))
	for i, f := range fields {
		args[i] = f
	}
	m.Called(args...)
	newFields := make(FieldSet, 0, len(m.fields)+len(fields))
	newFields = append(newFields, m.fields...)
	return &mockLoggerImpl{
		fields: append(newFields, fields...),
	}
}

func (m *mockLoggerImpl) sync() error {
	m.Called()
	return nil
}

func Test_baseLogger_Fatal(t *testing.T) {
	type fields struct {
		impl *mockLoggerImpl
	}
	type want struct {
		level Level
	}
	tests := []struct {
		name      string
		fields    fields
		want      want
		assertion func(*testing.T, loggerImpl, EventBuilder)
	}{
		{
			name:   "default",
			fields: fields{impl: &mockLoggerImpl{}},
			want:   want{level: LevelFatal},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			tt.fields.impl.On("clone").Return()

			e := l.Fatal().(*eventBuilder)

			tt.fields.impl.AssertExpectations(t)
			assert.NotEqual(t, tt.fields.impl, e.logger)
			assert.Equal(t, tt.want.level, e.level)
		})
	}
}

func Test_baseLogger_Error(t *testing.T) {
	type fields struct {
		impl *mockLoggerImpl
	}
	type want struct {
		level Level
	}
	tests := []struct {
		name      string
		fields    fields
		want      want
		assertion func(*testing.T, loggerImpl, EventBuilder)
	}{
		{
			name:   "default",
			fields: fields{impl: &mockLoggerImpl{}},
			want:   want{level: LevelError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			tt.fields.impl.On("clone").Return()

			e := l.Error().(*eventBuilder)

			tt.fields.impl.AssertExpectations(t)
			assert.NotEqual(t, tt.fields.impl, e.logger)
			assert.Equal(t, tt.want.level, e.level)
		})
	}
}

func Test_baseLogger_Info(t *testing.T) {
	type fields struct {
		impl *mockLoggerImpl
	}
	type want struct {
		level Level
	}
	tests := []struct {
		name      string
		fields    fields
		want      want
		assertion func(*testing.T, loggerImpl, EventBuilder)
	}{
		{
			name:   "default",
			fields: fields{impl: &mockLoggerImpl{}},
			want:   want{level: LevelInfo},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			tt.fields.impl.On("clone").Return()

			e := l.Info().(*eventBuilder)

			tt.fields.impl.AssertExpectations(t)
			assert.NotEqual(t, tt.fields.impl, e.logger)
			assert.Equal(t, tt.want.level, e.level)
		})
	}
}

func Test_baseLogger_Debug(t *testing.T) {
	type fields struct {
		impl *mockLoggerImpl
	}
	type want struct {
		level Level
	}
	tests := []struct {
		name      string
		fields    fields
		want      want
		assertion func(*testing.T, loggerImpl, EventBuilder)
	}{
		{
			name:   "default",
			fields: fields{impl: &mockLoggerImpl{}},
			want:   want{level: LevelDebug},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			tt.fields.impl.On("clone").Return()

			e := l.Debug().(*eventBuilder)

			tt.fields.impl.AssertExpectations(t)
			assert.NotEqual(t, tt.fields.impl, e.logger)
			assert.Equal(t, tt.want.level, e.level)
		})
	}
}

func Test_baseLogger_With(t *testing.T) {
	type fields struct {
		impl *mockLoggerImpl
	}
	type args struct {
		fields []Field
	}
	type want struct {
		fields []Field
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		// TODO: Add test cases.
		{
			name:   "no fields",
			fields: fields{impl: &mockLoggerImpl{}},
			args:   args{fields: []Field{}},
			want:   want{fields: []Field{}},
		},
		{
			name:   "single field",
			fields: fields{impl: &mockLoggerImpl{}},
			args:   args{fields: []Field{Int("f1", 123)}},
			want:   want{fields: []Field{Int("f1", 123)}},
		},
		{
			name:   "multiple fields",
			fields: fields{impl: &mockLoggerImpl{}},
			args:   args{fields: []Field{Int("f1", 123), Str("f2", "value"), Err("f3", errors.New("an error"))}},
			want:   want{fields: []Field{Int("f1", 123), Str("f2", "value"), Err("f3", errors.New("an error"))}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			origFields := tt.fields.impl.fields
			args := make([]any, len(tt.args.fields))
			for i, f := range tt.args.fields {
				args[i] = f
			}
			tt.fields.impl.On("with", args...).Return(mock.Anything)

			logger := l.With(tt.args.fields...).(*baseLogger)

			tt.fields.impl.AssertExpectations(t)
			assert.NotEqual(t, l, logger)
			assert.Equal(t, origFields, l.impl.(*mockLoggerImpl).fields)
			assert.ElementsMatch(t, tt.want.fields, logger.impl.(*mockLoggerImpl).fields)
		})
	}
}

func Test_baseLogger_Sync(t *testing.T) {
	type fields struct {
		impl *mockLoggerImpl
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "default",
			fields: fields{impl: &mockLoggerImpl{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			tt.fields.impl.On("sync").Return()

			l.Sync()

			tt.fields.impl.AssertExpectations(t)
		})
	}
}

func Test_baseLogger_WithErr(t *testing.T) {
	type fields struct {
		impl loggerImpl
	}
	type args struct {
		err error
	}
	type want struct {
		got Logger
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := &baseLogger{
				impl: tt.fields.impl,
			}
			got := l.WithErr(tt.args.err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
