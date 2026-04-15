package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
)

type LauncherOpts struct {
	MemProfileRate uint
	Timeout        uint
	OutputDir      string
}

type Launcher struct {
	T             *TestingT
	opts          LauncherOpts
	tempFactory   *servertest.TempFileFactory
	addrFactory   *servertest.ListenAddressFactory
	hmacKeyBase64 string
	binPathAgent  string
	binPathServer string
}

func NewLauncher(t *TestingT) *Launcher {
	l := &Launcher{
		T:           t,
		opts:        LauncherOpts{},
		tempFactory: servertest.NewTempFileFactory(t),
		addrFactory: servertest.NewListenAddressFactory(t),
	}

	l.generateHMACKey()

	return l
}

func (l *Launcher) Run() error {
	err := l.parseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %w", err)
	}

	err = os.MkdirAll(l.opts.OutputDir, 0o700)
	if err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err = l.compile(ctx)
	if err != nil {
		return fmt.Errorf("cannot compile binaries: %w", err)
	}

	return l.orchestrate(ctx)
}

func (l *Launcher) Cleanup() {
	l.tempFactory.RemoveAll()
	l.addrFactory.Clear()
}

func (l *Launcher) generateHMACKey() {
	buf := [32]byte{}
	rand.Read(buf[:])

	l.hmacKeyBase64 = base64.StdEncoding.EncodeToString(buf[:])
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

func (l *Launcher) compile(ctx context.Context) error {
	var err error

	l.binPathServer, err = l.compileCmd(ctx, "server")
	if err != nil {
		return fmt.Errorf("cannot compile server: %w", err)
	}

	l.binPathAgent, err = l.compileCmd(ctx, "agent")
	if err != nil {
		return fmt.Errorf("cannot compile agent: %w", err)
	}

	return nil
}

func (l *Launcher) orchestrate(baseCtx context.Context) error {
	dbCfg := servertest.LaunchEmbeddedPostgres(l.T, "pprof", "pprof", "pprof")

	addr := l.addrFactory.Get(0)

	l.T.Logf("running for up to %d seconds", l.opts.Timeout)

	ctx, cancel := context.WithTimeout(baseCtx, time.Duration(l.opts.Timeout)*time.Second)
	defer cancel()

	grp := new(errgroup.Group)

	grp.Go(func() error {
		return l.runServer(ctx, addr, dbCfg)
	})

	grp.Go(func() error {
		return l.runAgent(ctx, addr)
	})

	return grp.Wait()
}

func (l *Launcher) compileCmd(ctx context.Context, name string) (string, error) {
	l.T.Logf("compiling %s", name)

	outfile := l.tempFactory.Create(fmt.Sprintf("%s-", name))

	cmd := exec.CommandContext(ctx, "go", "build", "-o", outfile, fmt.Sprintf("./cmd/%s/", name))
	err := cmd.Run()

	return outfile, err
}

func (l *Launcher) runServer(ctx context.Context, addr string, dbCfg dbconfig.Config) error {
	return l.runCmd(ctx, "server", func(ctx context.Context, logfile io.Writer) *exec.Cmd {
		return createCmd(
			ctx,
			logfile,
			[]string{l.binPathServer},
			map[string]string{
				"ADDRESS":           addr,
				"SHUTDOWN_TIMEOUT":  "5", // 5 seconds
				"STORE_INTERVAL":    "5", // 5 seconds
				"FILE_STORAGE_PATH": l.tempFactory.Create("metric-dump-*"),
				"DATABASE_DSN":      dbCfg.DSN(),
				"KEY":               l.hmacKeyBase64,
				"AUDIT_FILE":        l.tempFactory.Create("metric-audit-*"),
				"PPROF_CPU_OUT":     filepath.Join(l.opts.OutputDir, "server.cpu.out"),
				"PPROF_MEM_OUT":     filepath.Join(l.opts.OutputDir, "server.mem.out"),
				"PPROF_MEM_RATE":    strconv.Itoa(int(l.opts.MemProfileRate)),
			},
		)
	})
}

func (l *Launcher) runAgent(ctx context.Context, addr string) error {
	time.Sleep(500 * time.Millisecond) // give server time to apply migrations

	return l.runCmd(ctx, "agent", func(ctx context.Context, logfile io.Writer) *exec.Cmd {
		return createCmd(
			ctx,
			logfile,
			[]string{l.binPathAgent},
			map[string]string{
				"ADDRESS":         fmt.Sprintf("http://%s", addr),
				"POLL_INTERVAL":   "1", // 1 second
				"REPORT_INTERVAL": "1", // 1 second
				"RATE_LIMIT":      strconv.Itoa(runtime.NumCPU()),
				"KEY":             l.hmacKeyBase64,
				"PPROF_CPU_OUT":   filepath.Join(l.opts.OutputDir, "agent.cpu.out"),
				"PPROF_MEM_OUT":   filepath.Join(l.opts.OutputDir, "agent.mem.out"),
				"PPROF_MEM_RATE":  strconv.Itoa(int(l.opts.MemProfileRate)),
			},
		)

	})
}

func (l *Launcher) runCmd(ctx context.Context, name string, cmdCreator func(context.Context, io.Writer) *exec.Cmd) error {
	logfile, err := os.Create(filepath.Join(l.opts.OutputDir, name+".log"))
	if err != nil {
		return fmt.Errorf("cannot create %s log file: %w", name, err)
	}
	defer logfile.Close()

	deadline, _ := ctx.Deadline()
	ctxCmd, cancelCmd := context.WithDeadline(context.Background(), deadline.Add(10*time.Second)) // extra time before sending SIGKILL
	defer cancelCmd()

	cmd := cmdCreator(ctxCmd, logfile)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("cannot start %s: %w", name, err)
	}
	l.T.Logf("%s started", name)

	go sigtermAfterContextIsDone(ctx, cmd)

	err = cmd.Wait()
	l.T.Logf("%s finished: %v", name, err)

	return err

}

func createCmd(ctx context.Context, logfile io.Writer, args []string, env map[string]string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	cmd.Stdout = logfile
	cmd.Stderr = logfile

	cmd.Env = os.Environ()

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	return cmd
}

func sigtermAfterContextIsDone(ctx context.Context, cmd *exec.Cmd) error {
	<-ctx.Done()

	if cmd.Process != nil {
		return cmd.Process.Signal(syscall.SIGTERM)
	}

	return nil
}
