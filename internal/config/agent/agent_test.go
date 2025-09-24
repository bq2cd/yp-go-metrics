package agent

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name      string
		args      args
		want      *Config
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty",
			args: args{},
			want: &Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "upstream url",
			args: args{opts: []Option{UpstreamURL("localhost:91")}},
			want: &Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:91"}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "poll interval",
			args: args{opts: []Option{PollInterval(5)}},
			want: &Config{PollInterval: 5 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "report interval",
			args: args{opts: []Option{ReportInterval(5)}},
			want: &Config{ReportInterval: 5 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple options",
			args: args{opts: []Option{
				UpstreamURL("http://example.com/test"),
				PollInterval(10),
				ReportInterval(5),
			}},
			want: &Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: "example.com", Path: "/test"},
				PollInterval:   10 * time.Second,
				ReportInterval: 5 * time.Second,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple options but some invalid",
			args: args{opts: []Option{
				UpstreamURL("http://example.com/test"),
				PollInterval(0),
				ReportInterval(5),
			}},
			want: nil,
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opts...)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpstreamURL(t *testing.T) {
	type args struct {
		upstreamURL string
	}
	type want struct {
		upstreamURL url.URL
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "emtpy",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "invalid",
			args:   args{upstreamURL: "http://localhost:91:invalid/123"},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "valid with empty config",
			args:   args{upstreamURL: "localhost:91"},
			want:   want{upstreamURL: url.URL{Scheme: "http", Host: "localhost:91"}},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.upstreamURL, c.UpstreamURL)
			},
		},
		{
			name:   "valid with empty config 2",
			args:   args{upstreamURL: "https://example.com:443"},
			want:   want{upstreamURL: url.URL{Scheme: "https", Host: "example.com:443"}},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.upstreamURL, c.UpstreamURL)
			},
		},
		{
			name:   "valid with preexisting config",
			args:   args{upstreamURL: "https://example.com:443"},
			want:   want{upstreamURL: url.URL{Scheme: "https", Host: "example.com:443"}},
			config: Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:8080"}},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.upstreamURL, c.UpstreamURL)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpstreamURL(tt.args.upstreamURL)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func Test_setInterval(t *testing.T) {
	type args struct {
		intervalSec uint
	}
	type target struct {
		interval time.Duration
	}
	type want struct {
		interval time.Duration
	}
	tests := []struct {
		name      string
		args      args
		target    target
		want      want
		assertion func(assert.TestingT, *target, error, want)
	}{
		{
			name: "zero",
			args: args{},
			want: want{},
			assertion: func(t assert.TestingT, target *target, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name: "positive",
			args: args{intervalSec: 55},
			want: want{interval: 55 * time.Second},
			assertion: func(t assert.TestingT, target *target, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, target.interval, want.interval)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setInterval(tt.args.intervalSec, &tt.target.interval)
			tt.assertion(t, &tt.target, err, tt.want)
		})
	}
}

func TestPollInterval(t *testing.T) {
	type args struct {
		intervalSec uint
	}
	type want struct {
		interval time.Duration
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.interval, c.PollInterval)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{PollInterval: 10 * time.Second},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.interval, c.PollInterval)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PollInterval(tt.args.intervalSec)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestReportInterval(t *testing.T) {
	type args struct {
		intervalSec uint
	}
	type want struct {
		interval time.Duration
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.interval, c.ReportInterval)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{PollInterval: 10 * time.Second},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.interval, c.ReportInterval)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ReportInterval(tt.args.intervalSec)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		UpstreamURL    url.URL
		PollInterval   time.Duration
		ReportInterval time.Duration
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty",
			fields: fields{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "zero poll interval",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				ReportInterval: 5 * time.Second,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "zero report interval",
			fields: fields{
				UpstreamURL:  url.URL{Host: "localhost:91"},
				PollInterval: 5 * time.Second,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "report interval < poll interval",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				PollInterval:   5 * time.Second,
				ReportInterval: 1 * time.Second,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "normal values",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				PollInterval:   5 * time.Second,
				ReportInterval: 10 * time.Second,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				UpstreamURL:    tt.fields.UpstreamURL,
				PollInterval:   tt.fields.PollInterval,
				ReportInterval: tt.fields.ReportInterval,
			}
			tt.assertion(t, c.Validate())
		})
	}
}
