package cli_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/cli"
	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type FakeConfig struct{}

var (
	FakeParseArgsFn = func(fs *flag.FlagSet, args []string, _ envparser.Parser) (FakeConfig, error) {
		err := fs.Parse(args)

		return FakeConfig{}, err
	}
	FakeLaunchProcessFn = func(ctx context.Context, _ log.Logger, _ FakeConfig) error {
		<-ctx.Done()

		return context.Cause(ctx)
	}
)

var (
	ErrAppShuttingDown = errors.New("app is shutting down")
)

func TestRun(t *testing.T) {
	tempFactory := servertest.NewTempFileFactory(t)
	t.Cleanup(tempFactory.RemoveAll)

	defaultTimeout := 20 * time.Millisecond

	type testcase struct {
		timeout   time.Duration
		args      []string
		env       map[string]string
		buildInfo cli.BuildInfo
		assertFn  func(*testing.T, error)
	}

	tests := map[string]testcase{
		"normal run": {
			args: []string{},
			assertFn: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrAppShuttingDown)
			},
		},
		"cpu profiling, invalid output path": {
			args: []string{"-pprof-cpu-out", "/tmp"},
			assertFn: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "cannot create cpu profile output")
			},
		},
		"mem profiling, invalid output path": {
			args: []string{"-pprof-mem-out", "/tmp"},
			assertFn: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "cannot create mem profile output")
			},
		},
		"mem profiling, invadid mem rate": {
			args: []string{"-pprof-mem-out", tempFactory.Create("pprof-mem-out-*"), "-pprof-mem-rate", "0"},
			assertFn: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "memory rate must be positive")
			},
		},
		"cpu profiling, valid output path": func() testcase {
			outCPU := tempFactory.Create("pprof-cpu-out-*")

			return testcase{
				args: []string{"-pprof-cpu-out", outCPU},
				assertFn: func(t *testing.T, err error) {
					require.ErrorIs(t, err, ErrAppShuttingDown)

					assertFilesExistAndNotEmpty(t, outCPU)
				},
			}
		}(),
		"mem profiling, valid output path and mem rate": func() testcase {
			outMem := tempFactory.Create("pprof-cpu-mem-*")

			return testcase{
				args: []string{"-pprof-mem-out", outMem, "-pprof-mem-rate", "1"},
				assertFn: func(t *testing.T, err error) {
					require.ErrorIs(t, err, ErrAppShuttingDown)

					assertFilesExistAndNotEmpty(t, outMem)
				},
			}
		}(),
		"cpu and mem profiling, valid output paths": func() testcase {
			outCPU := tempFactory.Create("pprof-cpu-out-*")
			outMem := tempFactory.Create("pprof-cpu-mem-*")

			return testcase{
				args: []string{"-pprof-cpu-out", outCPU, "-pprof-mem-out", outMem},
				assertFn: func(t *testing.T, err error) {
					require.ErrorIs(t, err, ErrAppShuttingDown)

					assertFilesExistAndNotEmpty(t, outCPU, outMem)
				},
			}
		}(),
		"cpu and mem profiling via env": func() testcase {
			outCPU := tempFactory.Create("pprof-cpu-out-*")
			outMem := tempFactory.Create("pprof-cpu-mem-*")

			return testcase{
				args: []string{},
				env: map[string]string{
					"PPROF_CPU_OUT":  outCPU,
					"PPROF_MEM_OUT":  outMem,
					"PPROF_MEM_RATE": "1",
				},
				assertFn: func(t *testing.T, err error) {
					require.ErrorIs(t, err, ErrAppShuttingDown)

					assertFilesExistAndNotEmpty(t, outCPU, outMem)
				},
			}
		}(),
		"build info overrides": {
			args: []string{},
			buildInfo: cli.BuildInfo{
				Version: "1.2.3",
				Date:    "now",
				Commit:  "abcdef0123",
			},
			assertFn: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrAppShuttingDown)
			},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

			app := cli.App[FakeConfig]{
				Name:          "fake-app",
				ParseArgs:     FakeParseArgsFn,
				LaunchProcess: FakeLaunchProcessFn,
				TerminalConfig: cli.TerminalConfig{
					Stdout: stdout,
					Stderr: stderr,
				},
				BuildInfo: tc.buildInfo,
			}

			logger := log.NewTestLogger()

			if tc.timeout == 0 {
				tc.timeout = defaultTimeout
			}

			overrideEnv(t, tc.env)

			ctx, cancel := context.WithTimeoutCause(t.Context(), tc.timeout, ErrAppShuttingDown)
			defer cancel()

			// Act
			err := app.Run(ctx, logger, tc.args)

			// Assert
			tc.assertFn(t, err)

			assertBuildInfoPrinted(t, tc.buildInfo, stdout.String())
		})
	}
}

func assertFilesExistAndNotEmpty(t *testing.T, files ...string) {
	for _, path := range files {
		if assert.FileExists(t, path) {
			stat, _ := os.Stat(path)
			assert.Positive(t, stat.Size())
		}
	}
}

func assertBuildInfoPrinted(t *testing.T, want cli.BuildInfo, stdout string) {
	buf := bytes.NewBuffer(nil)

	want.PrintTo(buf)

	assert.Equal(t, buf.String(), stdout)
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
