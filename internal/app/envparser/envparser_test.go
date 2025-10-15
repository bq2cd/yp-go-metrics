package envparser

import (
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(*testing.T, *parser)
	}{
		{
			name: "options must be nil",
			assertion: func(t *testing.T, got *parser) {
				assert.Nil(t, got.options)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, NewParser())
		})
	}
}

func TestNewParserWithOptions(t *testing.T) {
	type args struct {
		opts env.Options
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, *env.Options, *parser)
	}{
		{
			name: "empty options",
			args: args{opts: env.Options{}},
			assertion: func(t *testing.T, want *env.Options, got *parser) {
				assert.Equal(t, want, got.options)
			},
		},
		{
			name: "some options",
			args: args{opts: env.Options{TagName: "custom"}},
			assertion: func(t *testing.T, want *env.Options, got *parser) {
				assert.Equal(t, want, got.options)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, &tt.args.opts, NewParserWithOptions(tt.args.opts))
		})
	}
}

func Test_parser_Parse(t *testing.T) {
	type fields struct {
		options *env.Options
	}
	type args struct {
		setEnv func(*testing.T)
		v      any
	}
	type want struct {
		v any
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "default options",
			fields: fields{
				options: nil,
			},
			args: args{
				setEnv: func(t *testing.T) {
					t.Setenv("VAR1", "-15")
					t.Setenv("VAR2", "31")
					t.Setenv("VAR3", "dummy")
				},
				v: &struct {
					VarInt    int    `env:"VAR1"`
					VarUint   uint   `env:"VAR2"`
					VarString string `env:"VAR3"`
				}{},
			},
			want: want{
				v: &struct {
					VarInt    int    `env:"VAR1"`
					VarUint   uint   `env:"VAR2"`
					VarString string `env:"VAR3"`
				}{
					VarInt:    -15,
					VarUint:   31,
					VarString: "dummy",
				},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "with custom environment",
			fields: fields{
				options: &env.Options{
					Environment: map[string]string{"VAR1": "-12", "VAR2": "12", "VAR3": "a string"},
				},
			},
			args: args{
				setEnv: func(t *testing.T) {
					t.Setenv("VAR1", "-15")
					t.Setenv("VAR2", "31")
					t.Setenv("VAR3", "dummy string")
				},
				v: &struct {
					VarInt    int    `env:"VAR1"`
					VarUint   uint   `env:"VAR2"`
					VarString string `env:"VAR3"`
				}{},
			},
			want: want{
				v: &struct {
					VarInt    int    `env:"VAR1"`
					VarUint   uint   `env:"VAR2"`
					VarString string `env:"VAR3"`
				}{
					VarInt:    -12,
					VarUint:   12,
					VarString: "a string",
				},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "private fields are ignored",
			fields: fields{
				options: &env.Options{
					Environment: map[string]string{"VAR1": "-12", "VAR2": "12"},
				},
			},
			args: args{
				setEnv: func(t *testing.T) {},
				v: &struct {
					Var1 int  `env:"VAR1"`
					var2 uint `env:"VAR2"`
				}{},
			},
			want: want{
				v: &struct {
					Var1 int  `env:"VAR1"`
					var2 uint `env:"VAR2"`
				}{
					Var1: -12,
					var2: 0,
				},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "custom tag name",
			fields: fields{
				options: &env.Options{
					TagName:     "custom",
					Environment: map[string]string{"VAR1": "-12", "VAR2": "12"},
				},
			},
			args: args{
				setEnv: func(t *testing.T) {},
				v: &struct {
					Var1 int  `custom:"VAR1"`
					Var2 uint `env:"VAR2"`
				}{},
			},
			want: want{
				v: &struct {
					Var1 int  `custom:"VAR1"`
					Var2 uint `env:"VAR2"`
				}{
					Var1: -12,
					Var2: 0,
				},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "custom envvar prefix",
			fields: fields{
				options: &env.Options{
					Prefix:      "RRR_",
					Environment: map[string]string{"RRR_VAR1": "-12", "RRR_VAR2": "12"},
				},
			},
			args: args{
				setEnv: func(t *testing.T) {},
				v: &struct {
					Var1 int  `env:"VAR1"`
					Var2 uint `env:"VAR2"`
				}{},
			},
			want: want{
				v: &struct {
					Var1 int  `env:"VAR1"`
					Var2 uint `env:"VAR2"`
				}{
					Var1: -12,
					Var2: 12,
				},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parser{
				options: tt.fields.options,
			}
			tt.args.setEnv(t)
			tt.assertion(t, p.Parse(tt.args.v))
			assert.Equal(t, tt.want.v, tt.args.v)
		})
	}
}
