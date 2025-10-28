package retrymgr

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSleeper(t *testing.T) {
	assert.NotNil(t, NewSleeper())
}

func Test_sleeper_Sleep(t *testing.T) {
	type args struct {
		timeout time.Duration
		delay   time.Duration
	}
	type want struct {
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"normal run": {
			args: args{
				timeout: 100 * time.Millisecond,
				delay:   50 * time.Millisecond,
			},
			want: want{wantErr: func(t testing.TB, err error) {
				require.NoError(t, err)
			}},
		},
		"context canceled": {
			args: args{
				timeout: 50 * time.Millisecond,
				delay:   100 * time.Millisecond,
			},
			want: want{wantErr: func(t testing.TB, err error) {
				require.Errorf(t, err, "context canceled")
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.args.timeout)
			defer cancel()
			s := NewSleeper()

			err := s.Sleep(ctx, tt.args.delay)

			tt.want.wantErr(t, err)
		})
	}
}
