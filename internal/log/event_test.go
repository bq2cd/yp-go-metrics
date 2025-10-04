package log

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newEventBuilder(t *testing.T) {
	type args struct {
		logger *mockLoggerImpl
		level  Level
	}
	type want struct {
		level Level
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default",
			args: args{logger: &mockLoggerImpl{}, level: LevelInfo},
			want: want{level: LevelInfo},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEventBuilder(tt.args.logger, tt.args.level)
			assert.Equal(t, tt.args.logger, e.logger)
			assert.Equal(t, tt.want.level, e.level)
			assert.Len(t, e.fields, 0)
			assert.Greater(t, cap(e.fields), 0)
		})
	}
}

func Test_eventBuilder_With(t *testing.T) {
	type fields struct {
		logger *mockLoggerImpl
		level  Level
		fields FieldSet
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
		{
			name: "no fields",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{},
			},
			args: args{fields: FieldSet{}},
			want: want{fields: FieldSet{}},
		},
		{
			name: "single field",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{},
			},
			args: args{fields: FieldSet{Int("f1", 123)}},
			want: want{fields: FieldSet{Int("f1", 123)}},
		},
		{
			name: "multiple fields",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{},
			},
			args: args{fields: FieldSet{Int("f1", 123), Str("f2", "a value"), Err("f3", errors.New("an error"))}},
			want: want{fields: FieldSet{Int("f1", 123), Str("f2", "a value"), Err("f3", errors.New("an error"))}},
		},
		{
			name: "no fields with pre-existing",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f0", -123)},
			},
			args: args{fields: FieldSet{}},
			want: want{fields: FieldSet{Int("f0", -123)}},
		},
		{
			name: "single field with pre-existing",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f0", -123)},
			},
			args: args{fields: FieldSet{Int("f1", 123)}},
			want: want{fields: FieldSet{Int("f0", -123), Int("f1", 123)}},
		},
		{
			name: "multiple fields with pre-existing",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f0", -123), Dur("t0", 5*time.Second)},
			},
			args: args{fields: FieldSet{Int("f1", 123), Str("f3", "extra")}},
			want: want{fields: FieldSet{Int("f0", -123), Dur("t0", 5*time.Second), Int("f1", 123), Str("f3", "extra")}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventBuilder{
				logger: tt.fields.logger,
				level:  tt.fields.level,
				fields: tt.fields.fields,
			}
			origFields := tt.fields.fields

			got := e.With(tt.args.fields...).(*eventBuilder)

			tt.fields.logger.AssertNotCalled(t, "clone")
			assert.Equal(t, e.logger, got.logger)
			assert.Equal(t, origFields, e.fields)
			assert.ElementsMatch(t, tt.want.fields, got.fields)
		})
	}
}

func Test_eventBuilder_Msg(t *testing.T) {
	type fields struct {
		logger *mockLoggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantNotCalled bool
	}{
		{
			name: "empty event",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{},
			},
		},
		{
			name: "logging disabled",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  _levelDisabled,
				fields: FieldSet{Int("f1", 123)},
			},
			wantNotCalled: true,
		},
		{
			name: "single field",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f1", 123)},
			},
		},
		{
			name: "multiple fields",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f1", 123), Str("f2", "a value"), Err("f3", errors.New("an error"))},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventBuilder{
				logger: tt.fields.logger,
				level:  tt.fields.level,
				fields: tt.fields.fields,
			}
			args := make([]any, len(tt.fields.fields)+2)
			args[0] = tt.fields.level
			args[1] = tt.args.msg
			for i, f := range tt.fields.fields {
				args[i+2] = f
			}
			tt.fields.logger.On("log", args).Return()

			e.Msg(tt.args.msg)

			if tt.wantNotCalled {
				tt.fields.logger.AssertNotCalled(t, "log", args)
			} else {
				tt.fields.logger.AssertExpectations(t)
			}
		})
	}
}

func Test_eventBuilder_Send(t *testing.T) {
	type fields struct {
		logger *mockLoggerImpl
		level  Level
		fields FieldSet
	}
	tests := []struct {
		name          string
		fields        fields
		wantNotCalled bool
	}{
		{
			name: "empty event",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{},
			},
		},
		{
			name: "logging disabled",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  _levelDisabled,
				fields: FieldSet{Int("f1", 123)},
			},
			wantNotCalled: true,
		},
		{
			name: "single field",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f1", 123)},
			},
		},
		{
			name: "multiple fields",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{Int("f1", 123), Str("f2", "a value"), Err("f3", errors.New("an error"))},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventBuilder{
				logger: tt.fields.logger,
				level:  tt.fields.level,
				fields: tt.fields.fields,
			}
			args := make([]any, len(tt.fields.fields)+2)
			args[0] = tt.fields.level
			args[1] = ""
			for i, f := range tt.fields.fields {
				args[i+2] = f
			}
			tt.fields.logger.On("log", args).Return()

			e.Send()

			if tt.wantNotCalled {
				tt.fields.logger.AssertNotCalled(t, "log", args)
			} else {
				tt.fields.logger.AssertExpectations(t)
			}
		})
	}
}
