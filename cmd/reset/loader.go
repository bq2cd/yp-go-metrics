package main

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

// LoadedPackage represents a Go package after parsing source code with [go/parser]
// and type-checking it with [go/types].
type LoadedPackage struct {
	Name      string
	Dir       string
	Fset      *token.FileSet
	Syntax    []*ast.File
	TypesInfo *types.Info
}

// Loader scans for Go packages, parses their source code and type-checks them with [go/types]
// using `golang.org/x/tools/go/packages` as the means to achieve that.
// It designed to work with [pipeline] package and implements [pipeline.Source] interface.
type Loader struct {
	mu       sync.Mutex
	startDir string
	packages []*packages.Package
}

// NewLoader creates an instance of [Loader] with given start directory.
// Go packages are searched only in the start directory and below.
func NewLoader(startDir string) *Loader {
	l := &Loader{
		startDir: startDir,
	}

	return l
}

// Init is called by [pipeline.Source] stage runner on stage startup.
// For [Loader], it performs all heavy lifting of parsing and type-checking Go packages
// under given start directory; for that, it uses `golang.org/x/tools/go/packages`.
func (l *Loader) Init(ctx context.Context) error {
	var err error

	cfg := l.loaderConfig(ctx)

	l.packages, err = packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("cannot load packages from %s: %w", l.startDir, err)
	}

	slog.DebugContext(ctx, "loaded packages", slog.Int("count", len(l.packages)))

	return nil
}

// Produce is called by workers of [pipeline.Source] stage, in a loop, **concurrently**.
// It must consistently return `false` when there is no more data to produce.
// If an error is returned, the worker's loop is aborted.
// For [Loader], it simply returns next package from pre-populated (by [Loader.Init]) `packages` slice.
func (l *Loader) Produce(ctx context.Context) (*LoadedPackage, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.packages) == 0 {
		return nil, false, nil
	}

	pkg, err := l.popLastPackage()
	if err != nil {
		return nil, false, err
	}

	loaded := &LoadedPackage{
		Name:      pkg.Name,
		Dir:       pkg.Dir,
		Fset:      pkg.Fset,
		Syntax:    pkg.Syntax,
		TypesInfo: pkg.TypesInfo,
	}

	return loaded, true, nil
}

// Close is called by [pipeline.Source] stage runner on stage shutdown.
// For [Loader], it simply clears `packages` slice to allow GC to reclaim allocated memory.
func (l *Loader) Close(_ context.Context) error {
	l.packages = nil // allow GC to reclaim backing array

	return nil
}

func (l *Loader) loaderConfig(ctx context.Context) *packages.Config {
	cfg := &packages.Config{
		Mode:      packages.LoadSyntax,
		Context:   ctx,
		Dir:       l.startDir,
		ParseFile: l.parseFile,
		Tests:     false,
	}

	return cfg
}

func (l *Loader) parseFile(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	// ignore `reset.gen.go` (the file we generate)
	if filepath.Base(filename) == outputFilename {
		return nil, nil
	}

	return parser.ParseFile(fset, filename, src, parser.ParseComments|parser.AllErrors|parser.SkipObjectResolution)
}

func (l *Loader) popLastPackage() (*packages.Package, error) {
	var err error

	last := len(l.packages) - 1 // we check for length > 0 in [Produce]
	pkg := l.packages[last]

	l.packages = l.packages[:last]

	for i := range pkg.Errors {
		err = errors.Join(err, l.checkPackageError(pkg.Errors[i]))
	}

	if err != nil {
		return nil, fmt.Errorf("loader package error: %w", err)
	}

	if pkg.Module != nil && pkg.Module.Error != nil {
		return nil, fmt.Errorf("loader module error: %s", pkg.Module.Error.Err)
	}

	return pkg, nil
}

func (l *Loader) checkPackageError(perr packages.Error) error {
	switch perr.Kind {
	case packages.ListError:
		// hacky way to ignore `go list` errors related to `reset.gen.go` (the file we generate)
		if strings.Contains(perr.Msg, outputFilename) {
			return nil
		}
	}

	return perr
}
