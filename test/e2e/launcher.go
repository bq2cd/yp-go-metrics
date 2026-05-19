package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
)

// LauncherOpts defines [Launcher] options.
type LauncherOpts struct {
	Timeout      time.Duration
	AgentArgs    []string
	AgentEnv     map[string]string
	AgentOutput  io.Writer
	ServerArgs   []string
	ServerEnv    map[string]string
	ServerOutput io.Writer
}

// Launcher performs orchestration of agent and server processes.
type Launcher struct {
	T             servertest.TestingT
	opts          LauncherOpts
	tempFactory   *servertest.TempFileFactory
	addrFactory   *servertest.ListenAddressFactory
	selfDir       string
	dbConfig      dbconfig.Config
	binPathAgent  string
	binPathServer string
}

// NewLauncher creates an instance of [Launcher].
func NewLauncher(t servertest.TestingT, opts LauncherOpts) *Launcher {
	// Hack from https://stackoverflow.com/a/38644571 to get path to the current file.
	// We need this because in tests current working directory is set to the test package directory,
	// which breaks compiling of binaries that relies on path relative to the current working directory.
	_, file, _, _ := runtime.Caller(0)

	l := &Launcher{
		T:           t,
		opts:        opts,
		tempFactory: servertest.NewTempFileFactory(t),
		addrFactory: servertest.NewListenAddressFactory(t),
		selfDir:     filepath.Dir(file),
	}

	return l
}

// Run is the main entry point, that performs all the logic of compiling, launching
// and terminating of the server and agent processes.
func (l *Launcher) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := l.compile(ctx)
	if err != nil {
		return fmt.Errorf("cannot compile binaries: %w", err)
	}

	return l.orchestrate(ctx)
}

// Cleanup is responsible for removing all temporary files created during [Launcher.Run] execution.
func (l *Launcher) Cleanup() {
	l.tempFactory.RemoveAll()
	l.addrFactory.Clear()
}

// DBConfig exposes temporary database config to allow external processes to interact with the database.
func (l *Launcher) DBConfig() dbconfig.Config {
	return l.dbConfig
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
	l.dbConfig = servertest.LaunchEmbeddedPostgres(l.T, "e2e_user", "e2e_password", "e2e_db")

	addr := l.addrFactory.Get(0)

	l.T.Logf("running for up to %v", l.opts.Timeout)

	ctx, cancel := context.WithTimeout(baseCtx, l.opts.Timeout)
	defer cancel()

	grp := new(errgroup.Group)

	grp.Go(func() error {
		return l.runServer(ctx, addr)
	})

	grp.Go(func() error {
		return l.runAgent(ctx, addr)
	})

	return grp.Wait()
}

func (l *Launcher) compileCmd(ctx context.Context, name string) (string, error) {
	outfile := l.tempFactory.Create(fmt.Sprintf("%s-", name))

	l.T.Logf("compiling %s (outfile=%s)", name, outfile)

	srcDir := filepath.Join(l.selfDir, "..", "..", "cmd", name)

	buf := bytes.NewBuffer(nil)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", outfile, srcDir+"/")
	cmd.Stdout = buf
	cmd.Stderr = buf

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%w (stdout+stderr=%s)", err, buf.String())
	}

	return outfile, err
}

func (l *Launcher) runServer(ctx context.Context, addr string) error {
	env := maps.Clone(l.opts.ServerEnv)
	if env == nil {
		env = make(map[string]string)
	}

	maps.Insert(env, maps.All(map[string]string{
		"ADDRESS":          addr,
		"SHUTDOWN_TIMEOUT": "5", // 5 seconds
		"DATABASE_DSN":     l.dbConfig.DSN(),
	}))

	return l.runCmd(ctx, "server", func(ctx context.Context) *exec.Cmd {
		l.T.Logf("server will be listening at %s", addr)

		return createCmd(
			ctx,
			l.opts.ServerOutput,
			append([]string{l.binPathServer}, l.opts.ServerArgs...),
			env,
		)
	})
}

func (l *Launcher) runAgent(ctx context.Context, addr string) error {
	time.Sleep(500 * time.Millisecond) // give server time to apply migrations

	env := maps.Clone(l.opts.AgentEnv)
	if env == nil {
		env = make(map[string]string)
	}

	maps.Insert(env, maps.All(map[string]string{
		"ADDRESS": fmt.Sprintf("http://%s", addr),
	}))

	return l.runCmd(ctx, "agent", func(ctx context.Context) *exec.Cmd {
		return createCmd(
			ctx,
			l.opts.AgentOutput,
			append([]string{l.binPathAgent}, l.opts.AgentArgs...),
			env,
		)
	})
}

func (l *Launcher) runCmd(ctx context.Context, name string, cmdCreator func(context.Context) *exec.Cmd) error {
	deadline, _ := ctx.Deadline()
	ctxCmd, cancelCmd := context.WithDeadline(context.Background(), deadline.Add(10*time.Second)) // extra time before sending SIGKILL
	defer cancelCmd()

	cmd := cmdCreator(ctxCmd)

	err := cmd.Start()
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
