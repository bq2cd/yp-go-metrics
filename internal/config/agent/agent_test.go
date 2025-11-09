package agent

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:      "empty",
			args:      args{},
			want:      &Config{},
			assertion: assert.NoError,
		},
		{
			name:      "upstream url",
			args:      args{opts: []Option{UpstreamURL("localhost:91")}},
			want:      &Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:91"}},
			assertion: assert.NoError,
		},
		{
			name:      "poll interval",
			args:      args{opts: []Option{PollInterval(5)}},
			want:      &Config{PollInterval: 5 * time.Second},
			assertion: assert.NoError,
		},
		{
			name:      "report interval",
			args:      args{opts: []Option{ReportInterval(5)}},
			want:      &Config{ReportInterval: 5 * time.Second},
			assertion: assert.NoError,
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
			assertion: assert.NoError,
		},
		{
			name: "multiple options but some invalid",
			args: args{opts: []Option{
				UpstreamURL("http://example.com/test"),
				PollInterval(0),
				ReportInterval(5),
			}},
			want:      nil,
			assertion: assert.Error,
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
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "empty",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		{
			name:   "invalid",
			args:   args{upstreamURL: "http://localhost:91:invalid/123"},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		{
			name:   "valid with empty config",
			args:   args{upstreamURL: "localhost:91"},
			want:   want{upstreamURL: url.URL{Scheme: "http", Host: "localhost:91"}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.upstreamURL, c.UpstreamURL)
			},
		},
		{
			name:   "valid with empty config 2",
			args:   args{upstreamURL: "https://example.com:443"},
			want:   want{upstreamURL: url.URL{Scheme: "https", Host: "example.com:443"}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.upstreamURL, c.UpstreamURL)
			},
		},
		{
			name:   "valid with preexisting config",
			args:   args{upstreamURL: "https://example.com:443"},
			want:   want{upstreamURL: url.URL{Scheme: "https", Host: "example.com:443"}},
			config: Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:8080"}},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
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
		assertion func(*testing.T, *target, error, want)
	}{
		{
			name: "zero",
			args: args{},
			want: want{},
			assertion: func(t *testing.T, target *target, err error, want want) {
				require.Error(t, err)
			},
		},
		{
			name: "positive",
			args: args{intervalSec: 55},
			want: want{interval: 55 * time.Second},
			assertion: func(t *testing.T, target *target, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.interval, target.interval)
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
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.interval, c.PollInterval)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{PollInterval: 10 * time.Second},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
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
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.interval, c.ReportInterval)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{PollInterval: 10 * time.Second},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
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
		HMACSecretKey  []byte
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "empty",
			fields:    fields{},
			assertion: assert.Error,
		},
		{
			name: "zero poll interval",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				ReportInterval: 5 * time.Second,
			},
			assertion: assert.Error,
		},
		{
			name: "zero report interval",
			fields: fields{
				UpstreamURL:  url.URL{Host: "localhost:91"},
				PollInterval: 5 * time.Second,
			},
			assertion: assert.Error,
		},
		{
			name: "report interval < poll interval",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				PollInterval:   5 * time.Second,
				ReportInterval: 1 * time.Second,
			},
			assertion: assert.Error,
		},
		{
			name: "normal values",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				PollInterval:   5 * time.Second,
				ReportInterval: 10 * time.Second,
			},
			assertion: assert.NoError,
		},
		{
			name: "with secret key",
			fields: fields{
				UpstreamURL:    url.URL{Host: "localhost:91"},
				PollInterval:   5 * time.Second,
				ReportInterval: 10 * time.Second,
				HMACSecretKey:  []byte(`123`),
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				UpstreamURL:    tt.fields.UpstreamURL,
				PollInterval:   tt.fields.PollInterval,
				ReportInterval: tt.fields.ReportInterval,
				HMACSecretKey:  tt.fields.HMACSecretKey,
			}
			tt.assertion(t, c.Validate())
		})
	}
}

func TestHMACSecretKey(t *testing.T) {
	type args struct {
		key string
	}
	type want struct {
		got []byte
	}
	type testcase struct {
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
		"empty key is okay": {
			args:   args{key: ""},
			want:   want{got: nil},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"non-empty plain-text key is accepted": {
			args:   args{key: `123`},
			want:   want{got: []byte(`123`)},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"non-empty base64-encoded key is accepted": {
			args:   args{key: "MTIz"},
			want:   want{got: []byte(`123`)},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"existing key can be overwritten": {
			args:   args{key: "MTIz"},
			want:   want{got: []byte(`123`)},
			config: Config{HMACSecretKey: []byte(`something`)},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"existing key can be overwritten with empty value": {
			args:   args{key: ""},
			want:   want{got: nil},
			config: Config{HMACSecretKey: []byte(`something`)},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &tt.config
			err := HMACSecretKey(tt.args.key)(c)
			tt.assertion(t, c, err, tt.want)
			assert.Equal(t, tt.want.got, c.HMACSecretKey)
		})
	}
}

func TestSenderPoolSize(t *testing.T) {
	type args struct {
		size uint
	}
	type want struct {
		size    uint
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		config Config
		args   args
		want   want
	}
	tests := map[string]testcase{
		"default value is zero": {
			config: Config{},
			args:   args{},
			want: want{
				size: 0,
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"new value is set": {
			config: Config{},
			args:   args{size: 15},
			want: want{
				size: 15,
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"old value is overridden": {
			config: Config{SenderPoolSize: 5},
			args:   args{size: 15},
			want: want{
				size: 15,
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &tt.config
			err := SenderPoolSize(tt.args.size)(c)
			tt.want.wantErr(t, err)
			assert.Equal(t, tt.want.size, c.SenderPoolSize)
		})
	}
}
