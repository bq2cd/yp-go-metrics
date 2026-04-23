// Binary goose provides a thin CLI wrapper on top of `github.com/pressly/goose` library
// to perform SQL migrations management. This approach is documented in the official documentation
// [here](https://github.com/pressly/goose/blob/main/examples/go-migrations/main.go).
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/bq2cd/yp-go-metrics/internal/app/cli"
	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/internal/app/logger"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

const (
	databaseDriver = "pgx"
)

type cliOptions struct {
	MigrationsDir string `env:"MIGRATIONS_DIR"`
	DatabaseURL   string `env:"DATABASE_DSN"`
	GooseCommand  string
	GooseArgs     []string
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (cliOptions, error) {
	var opts cliOptions

	fs.StringVar(&opts.MigrationsDir, "dir", ".", "directory with migration files")
	fs.StringVar(&opts.DatabaseURL, "d", "postgres://", "database connection string (only PostgreSQL is supported)")

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, fmt.Errorf("invalid args: %w", err)
	}
	if err := envParser.Parse(&opts); err != nil {
		return cliOptions{}, fmt.Errorf("invalid env vars: %w", err)
	}

	extraArgs := fs.Args()
	if len(extraArgs) == 0 {
		fs.Usage()
		return cliOptions{}, fmt.Errorf("must specify goose command (e.g. create, up, status, ...)")
	}
	opts.GooseCommand = extraArgs[0]
	opts.GooseArgs = extraArgs[1:]

	return opts, nil
}

func launchProcess(ctx context.Context, logger log.Logger, opts cliOptions) (errFinal error) {
	db, err := goose.OpenDBWithDriver(databaseDriver, opts.DatabaseURL)
	if err != nil {
		errFinal = fmt.Errorf("goose: failed to open DB: %w", err)
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			errFinal = fmt.Errorf("goose: failed to close DB: %w", err)
		}
	}()

	if err := goose.RunContext(ctx, opts.GooseCommand, db, opts.MigrationsDir, opts.GooseArgs...); err != nil {
		errFinal = fmt.Errorf("goose: %v: %w", opts.GooseCommand, err)
	}

	return
}

func run(ctx context.Context, args []string, stderr io.Writer) error {
	app := cli.App[cliOptions]{
		Name:          "goose",
		ParseArgs:     parseArgs,
		LaunchProcess: launchProcess,
	}
	return app.Run(ctx, logger.NewDevelopment(), args, stderr)
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Stderr)
	if err != nil {
		stdlog.Fatalln(err)
	}
}
