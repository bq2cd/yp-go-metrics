package logger

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fileCapturer struct {
	orig   *os.File
	setFn  func(w *os.File) *os.File
	writer *os.File
	reader *os.File
}

func (f *fileCapturer) Set() error {
	var err error
	f.reader, f.writer, err = os.Pipe()
	if err != nil {
		return err
	}
	f.orig = f.setFn(f.writer)
	return nil
}

func (f *fileCapturer) Restore() {
	f.setFn(f.orig)
}

func (f *fileCapturer) ReadAll() (string, error) {
	err := f.writer.Close()
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(f.reader)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

type captureRunner struct {
	t *testing.T
}

func (r *captureRunner) setupStdout() *fileCapturer {
	stdout := &fileCapturer{
		setFn: func(w *os.File) *os.File {
			orig := os.Stdout
			os.Stdout = w
			return orig
		},
	}
	require.NoError(r.t, stdout.Set())
	return stdout
}

func (r *captureRunner) setupStderr() *fileCapturer {
	stderr := &fileCapturer{
		setFn: func(w *os.File) *os.File {
			orig := os.Stderr
			os.Stderr = w
			return orig
		},
	}
	require.NoError(r.t, stderr.Set())
	return stderr
}

func (r *captureRunner) readCaptured(f *fileCapturer) string {
	out, err := f.ReadAll()
	require.NoError(r.t, err)
	return out
}

func (r *captureRunner) Run(targetFn func(*testing.T)) (string, string) {
	stdout := r.setupStdout()
	defer stdout.Restore()

	stderr := r.setupStderr()
	defer stdout.Restore()

	targetFn(r.t)

	return r.readCaptured(stdout), r.readCaptured(stderr)
}

func TestNewProduction(t *testing.T) {
	r := &captureRunner{t}

	_, stderr := r.Run(func(t *testing.T) {
		logger := NewProduction()
		logger.Info().Bool("production", true).Msg("mic test")
	})

	var got map[string]any
	err := json.Unmarshal([]byte(stderr), &got)
	require.NoError(t, err)
	delete(got, "ts")
	delete(got, "caller")

	assert.Equal(t, map[string]any{
		"level":      "info",
		"production": true,
		"msg":        "mic test",
	}, got)
}

func TestNewDevelopment(t *testing.T) {
	r := &captureRunner{t}

	stdout, _ := r.Run(func(t *testing.T) {
		logger := NewDevelopment()
		logger.Debug().Str("not_production_key", "not-production-value").Msg("not production message")
	})

	assert.Contains(t, stdout, " DBG ")
	assert.Contains(t, stdout, "not_production_key=")
	assert.Contains(t, stdout, "not-production-value")
	assert.Contains(t, stdout, " not production message ")
}
