package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopLogger(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(*testing.T, Logger)
	}{
		{
			name: "default",
			assertion: func(t *testing.T, got Logger) {
				assert.IsType(t, &baseLogger{}, got)
				l := got.(*baseLogger)
				assert.IsType(t, &noopLogger{}, l.impl)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, NewNoopLogger())
		})
	}
}

func Test_noopLogger_log(t *testing.T) {
	type args struct {
		lvl    Level
		msg    string
		fields []Field
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "noop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &noopLogger{}
			l.log(tt.args.lvl, tt.args.msg, tt.args.fields...)
		})
	}
}

func Test_noopLogger_clone(t *testing.T) {
	tests := []struct {
		name      string
		logger    *noopLogger
		assertion func(*testing.T, *noopLogger, loggerImpl)
	}{
		{
			name:   "returns the same",
			logger: &noopLogger{},
			assertion: func(t *testing.T, want *noopLogger, got loggerImpl) {
				require.IsType(t, &noopLogger{}, got)
				assert.Equal(t, want, got.(*noopLogger))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.logger, tt.logger.clone())
		})
	}
}

func Test_noopLogger_with(t *testing.T) {
	type args struct {
		fields []Field
	}
	tests := []struct {
		name      string
		logger    *noopLogger
		args      args
		assertion func(*testing.T, *noopLogger, loggerImpl)
	}{
		{
			name:   "returns the same",
			logger: &noopLogger{},
			assertion: func(t *testing.T, want *noopLogger, got loggerImpl) {
				require.IsType(t, &noopLogger{}, got)
				assert.Equal(t, want, got.(*noopLogger))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.logger, tt.logger.with(tt.args.fields...))
		})
	}
}

func Test_noopLogger_sync(t *testing.T) {
	tests := []struct {
		name      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &noopLogger{}
			tt.assertion(t, l.sync())
		})
	}
}
