package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	"github.com/bq2cd/yp-go-metrics/test/e2e"
)

// LauncherOpts defines [Launcher] options, settable from CLI flags.
type LauncherOpts struct {
	MemProfileRate uint
	Timeout        uint
	OutputDir      string
}

// Launcher performs orchestration of agent and server processes.
// It is responsible for enabling CPU and memory profiling and
// graceful shutdown of the launched processes.
type Launcher struct {
	T             *TestingT
	opts          LauncherOpts
	tempFactory   *servertest.TempFileFactory
	hmacKeyBase64 string
}

// NewLauncher creates an instance of [Launcher].
func NewLauncher(t *TestingT) (*Launcher, error) {
	l := &Launcher{
		T:           t,
		opts:        LauncherOpts{},
		tempFactory: servertest.NewTempFileFactory(t),
	}

	err := l.generateHMACKey()

	return l, err
}

// Run is the main entry point, that performs all the logic of compiling, launching
// and terminating of the server and agent processes.
func (l *Launcher) Run() error {
	err := l.parseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %w", err)
	}

	err = os.MkdirAll(l.opts.OutputDir, 0o700)
	if err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	serverLog, err := os.Create(filepath.Join(l.opts.OutputDir, "server.log"))
	if err != nil {
		return fmt.Errorf("cannot create server.log: %w", err)
	}

	defer serverLog.Close()

	agentLog, err := os.Create(filepath.Join(l.opts.OutputDir, "agent.log"))
	if err != nil {
		return fmt.Errorf("cannot create agent.log: %w", err)
	}

	defer agentLog.Close()

	launcher := e2e.NewLauncher(l.T, e2e.LauncherOpts{
		Timeout:      time.Duration(l.opts.Timeout) * time.Second,
		AgentArgs:    []string{},
		AgentEnv:     l.getAgentEnv(),
		AgentOutput:  agentLog,
		ServerArgs:   []string{},
		ServerEnv:    l.getServerEnv(),
		ServerOutput: serverLog,
	})

	defer launcher.Cleanup()

	return launcher.Run()
}

// Cleanup is responsible for removing all temporary files created during [Launcher.Run] execution.
func (l *Launcher) Cleanup() {
	l.tempFactory.RemoveAll()
}

func (l *Launcher) generateHMACKey() error {
	buf := [32]byte{}

	// As per documentation to [rand.Read], it never returns an error but rather crashes the program irrecoverably.
	// However, it is better to err on the safe side here :).
	_, err := rand.Read(buf[:])
	if err != nil {
		return fmt.Errorf("crypto/rand failed: %v", err)
	}

	l.hmacKeyBase64 = base64.StdEncoding.EncodeToString(buf[:])

	return nil
}

func (l *Launcher) parseArgs() error {
	fs := flag.NewFlagSet("pproftest", flag.ExitOnError)

	fs.UintVar(&l.opts.Timeout, "t", 60, "total time (in seconds) to run profiling)")
	fs.UintVar(&l.opts.MemProfileRate, "m", 1<<10, "profile only allocation larger than N bytes (see runtime.MemProfileRate for details)")
	fs.StringVar(&l.opts.OutputDir, "d", "", "output directory for generated pprof profiles")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if l.opts.OutputDir == "" {
		return fmt.Errorf("output directory is required (use -d to specify)")
	}

	stat, err := os.Stat(l.opts.OutputDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if stat != nil && !stat.IsDir() {
		return fmt.Errorf("output directory points to a file! (%v)", stat)
	}

	return nil
}

func (l *Launcher) getServerEnv() map[string]string {
	return map[string]string{
		"STORE_INTERVAL":    "5", // 5 seconds
		"FILE_STORAGE_PATH": l.tempFactory.Create("metric-dump-*"),
		"KEY":               l.hmacKeyBase64,
		"AUDIT_FILE":        l.tempFactory.Create("metric-audit-*"),
		"PPROF_CPU_OUT":     filepath.Join(l.opts.OutputDir, "server.cpu.out"),
		"PPROF_MEM_OUT":     filepath.Join(l.opts.OutputDir, "server.mem.out"),
		"PPROF_MEM_RATE":    strconv.Itoa(int(l.opts.MemProfileRate)),
	}
}

func (l *Launcher) getAgentEnv() map[string]string {
	return map[string]string{
		"POLL_INTERVAL":   "1", // 1 second
		"REPORT_INTERVAL": "1", // 1 second
		"RATE_LIMIT":      strconv.Itoa(runtime.NumCPU()),
		"KEY":             l.hmacKeyBase64,
		"PPROF_CPU_OUT":   filepath.Join(l.opts.OutputDir, "agent.cpu.out"),
		"PPROF_MEM_OUT":   filepath.Join(l.opts.OutputDir, "agent.mem.out"),
		"PPROF_MEM_RATE":  strconv.Itoa(int(l.opts.MemProfileRate)),
	}
}
