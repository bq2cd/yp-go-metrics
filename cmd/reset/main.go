// Binary reset searches the project for structs that have `// generate:reset` comment.
// For all such structs, an implementation of `Reset()` method is generated and placed into `reset.gen.go` file
// inside the struct's package directory.
// The `Reset()` method contains logic of clearing struct's fields according to the following rules:
// - primitive types are assigned their zero values (e.g. `0` for `int`, `false` for `bool`).
// - slices are reset to zero length (e.g. `slice[:0]`).
// - maps are cleared (e.g. `clear(map)`).
// - nested structs implementing `interface { Reset() }` have their `Reset()` method called.
// - non-nil pointers are dereferenced and the above logic applies to the values they point to.
// - any unknown structs (not implementing `Reset()` method) are skipped.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bq2cd/yp-go-metrics/cmd/reset/pipeline"
	"github.com/bq2cd/yp-go-metrics/cmd/reset/rungroup"
)

const (
	magicComment   = "generate:reset"
	outputFilename = "reset.gen.go"
)

// Config contains main program options, such as start directory, etc.
type Config struct {
	StartDir string
	Debug    bool
}

func main() {
	var (
		cfg Config
		err error
	)

	err = initConfig(&cfg)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = run(ctx, cfg)
	if err != nil {
		log.Fatalln(err)
	}
}

func initConfig(cfg *Config) error {
	flag.BoolVar(&cfg.Debug, "d", false, "enable debug logging")

	flag.Parse()

	startDir := flag.Arg(0)
	if startDir == "" {
		startDir = "."
	}

	startDirAbs, err := filepath.Abs(startDir)
	if err != nil {
		return fmt.Errorf("cannot derive absolute path from %s: %w", startDir, err)
	}

	cfg.StartDir = startDirAbs

	return nil
}

func run(baseCtx context.Context, cfg Config) error {
	initLogger(cfg.Debug)

	grp, ctx := rungroup.New(baseCtx)

	loader := pipeline.NewSource("loader", NewLoader(cfg.StartDir), pipeline.DefaultConfig())
	analyzer := pipeline.NewStep("analyzer", loader.Out(), NewAnalyzer(), pipeline.DefaultConfig())
	renderer := pipeline.NewSink("renderer", analyzer.Out(), NewRenderer(os.Stderr), pipeline.DefaultConfig())

	grp.Go(ctx, loader, analyzer, renderer)

	return grp.Wait()
}

func initLogger(debugEnabled bool) {
	level := slog.LevelInfo
	if debugEnabled {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)
}
