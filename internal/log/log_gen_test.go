package log

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	type args struct {
		key   string
		value bool
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: true},
			want: Field{Key: "key", Type: FieldTypeBool, Value: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Bool(tt.args.key, tt.args.value))
		})
	}
}

func TestInt(t *testing.T) {
	type args struct {
		key   string
		value int
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: -123},
			want: Field{Key: "key", Type: FieldTypeInt, Value: -123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Int(tt.args.key, tt.args.value))
		})
	}
}

func TestFloat(t *testing.T) {
	type args struct {
		key   string
		value float64
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: -123.456},
			want: Field{Key: "key", Type: FieldTypeFloat, Value: -123.456},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Float(tt.args.key, tt.args.value))
		})
	}
}

func TestStr(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: "value"},
			want: Field{Key: "key", Type: FieldTypeStr, Value: "value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Str(tt.args.key, tt.args.value))
		})
	}
}

func TestErr(t *testing.T) {
	type args struct {
		key   string
		value error
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: errors.New("an error")},
			want: Field{Key: "key", Type: FieldTypeErr, Value: errors.New("an error")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Err(tt.args.key, tt.args.value))
		})
	}
}

func TestDur(t *testing.T) {
	type args struct {
		key   string
		value time.Duration
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: 5 * time.Second},
			want: Field{Key: "key", Type: FieldTypeDur, Value: 5 * time.Second},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Dur(tt.args.key, tt.args.value))
		})
	}
}

func TestTime(t *testing.T) {
	type args struct {
		key   string
		value time.Time
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: time.Date(1925, time.April, 12, 0, 1, 2, 0, time.UTC)},
			want: Field{Key: "key", Type: FieldTypeTime, Value: time.Date(1925, time.April, 12, 0, 1, 2, 0, time.UTC)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Time(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Bool(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: true},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeBool, Value: true}},
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
			assert.Equal(t, tt.want, e.Bool(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Int(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: 123},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeInt, Value: 123}},
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
			assert.Equal(t, tt.want, e.Int(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Float(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: 0.05},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeFloat, Value: 0.05}},
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
			assert.Equal(t, tt.want, e.Float(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Str(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: "value"},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeStr, Value: "value"}},
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
			assert.Equal(t, tt.want, e.Str(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Err(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: errors.New("an error")},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeErr, Value: errors.New("an error")}},
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
			assert.Equal(t, tt.want, e.Err(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Dur(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: 5 * time.Second},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeDur, Value: 5 * time.Second}},
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
			assert.Equal(t, tt.want, e.Dur(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Time(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: time.Date(1925, time.April, 12, 0, 1, 2, 0, time.UTC)},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeTime, Value: time.Date(1925, time.April, 12, 0, 1, 2, 0, time.UTC)}},
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
			assert.Equal(t, tt.want, e.Time(tt.args.key, tt.args.value))
		})
	}
}

func TestAny(t *testing.T) {
	type args struct {
		key   string
		value any
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{
			name: "default",
			args: args{key: "key", value: map[string][]int{"k1": {1, 2, 3}}},
			want: Field{Key: "key", Type: FieldTypeAny, Value: map[string][]int{"k1": {1, 2, 3}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Any(tt.args.key, tt.args.value))
		})
	}
}

func Test_eventBuilder_Any(t *testing.T) {
	type fields struct {
		logger loggerImpl
		level  Level
		fields FieldSet
	}
	type args struct {
		key   string
		value any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   EventBuilder
	}{
		{
			name: "default",
			fields: fields{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: make(FieldSet, 0),
			},
			args: args{key: "key", value: map[string][]int{"k1": {1, 2, 3}}},
			want: &eventBuilder{
				logger: &mockLoggerImpl{},
				level:  LevelInfo,
				fields: FieldSet{{Key: "key", Type: FieldTypeAny, Value: map[string][]int{"k1": {1, 2, 3}}}},
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
			assert.Equal(t, tt.want, e.Any(tt.args.key, tt.args.value))
		})
	}
}
