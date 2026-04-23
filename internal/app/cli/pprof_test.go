package cli

import (
	"math"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

const (
	initialMemProfileRate = math.MaxInt
)

func TestProfiler_MaybeStartProfiling(t *testing.T) {
	tempFactory := servertest.NewTempFileFactory(t)
	t.Cleanup(tempFactory.RemoveAll)

	type testcase struct {
		opts     profilerOpts
		env      map[string]string
		assertFn func(*testing.T, error)
	}

	tests := map[string]testcase{
		"memory profiling is disabled if no output path": {
			opts: profilerOpts{},
			assertFn: func(t *testing.T, err error) {
				require.NoError(t, err)
				assert.Zero(t, runtime.MemProfileRate)
			},
		},
		"memory profiling is enabled if valid output path": {
			opts: profilerOpts{
				OutPathMem: tempFactory.Create("pprof-out-mem-*"),
				MemRate:    123,
			},
			assertFn: func(t *testing.T, err error) {
				require.NoError(t, err)
				assert.Equal(t, 123, runtime.MemProfileRate)
			},
		},
		"env vars override initial values": {
			opts: profilerOpts{
				OutPathMem: tempFactory.Create("pprof-out-mem-*"),
				MemRate:    123,
			},
			env: map[string]string{
				"PPROF_MEM_RATE": "456",
			},
			assertFn: func(t *testing.T, err error) {
				require.NoError(t, err)
				assert.Equal(t, 456, runtime.MemProfileRate)
			},
		},
		"invalid env vars cause error": {
			opts: profilerOpts{
				OutPathMem: tempFactory.Create("pprof-out-mem-*"),
				MemRate:    123,
			},
			env: map[string]string{
				"PPROF_MEM_RATE": "-123",
			},
			assertFn: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "cannot parse env vars")
				assert.Equal(t, initialMemProfileRate, runtime.MemProfileRate)
			},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			logger := log.NewTestLogger()

			profiler := newProfiler(logger)
			profiler.opts = tc.opts

			overrideEnv(t, tc.env)

			// start each test with a known value to eliminate
			// side effects from other tests
			runtime.MemProfileRate = initialMemProfileRate

			// Act
			stop, err := profiler.MaybeStartProfiling()
			stop()

			// Assert
			tc.assertFn(t, err)
		})
	}
}

func overrideEnv(t *testing.T, env map[string]string) {
	for k, v := range env {
		err := os.Setenv(k, v)
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for k := range env {
			err := os.Unsetenv(k)
			require.NoError(t, err)
		}
	})
}
