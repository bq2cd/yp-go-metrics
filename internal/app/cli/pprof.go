package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

const (
	profilerDefaultMemRate = 512 * 1024 // equivalent to default [runtime.MemProfileRate] value
)

var (
	// ErrProfilerMemRateNotPositive is returned when provided memory rate is zero (which disables profiling).
	ErrProfilerMemRateNotPositive = errors.New("memory rate must be positive and greater than zero")
)

type profilerStopFunc func()

type profilerOpts struct {
	OutPathCPU string `env:"PPROF_CPU_OUT"`
	OutPathMem string `env:"PPROF_MEM_OUT"`
	MemRate    uint   `env:"PPROF_MEM_RATE"`
}

type profiler struct {
	logger  log.Logger
	opts    profilerOpts
	stopFns []profilerStopFunc
}

func newProfiler(logger log.Logger) *profiler {
	return &profiler{
		logger:  logger,
		opts:    profilerOpts{},
		stopFns: make([]profilerStopFunc, 0),
	}
}

// AddProfilingArgs adds CLI flags for profiling configuration to the provided flag set.
func (p *profiler) AddProfilingArgs(fs *flag.FlagSet) {
	fs.StringVar(&p.opts.OutPathCPU, "pprof-cpu-out", "", "path to cpu profile output")
	fs.StringVar(&p.opts.OutPathMem, "pprof-mem-out", "", "path to memory (heap) profile output")
	fs.UintVar(&p.opts.MemRate, "pprof-mem-rate", profilerDefaultMemRate, "profile memory allocation every N bytes (must be > 0)")
}

// MaybeStartProfiling parses environment variables and conditionally enables CPU and/or memory profiling.
// It returns a `stop` function that must be called before process termination in order for the profiling data
// to be collected properly.
// Typically, `stop` function is called at the end of `main` function.
func (p *profiler) MaybeStartProfiling() (profilerStopFunc, error) {
	stopFn := p.stopProfiling

	err := p.parseEnv()
	if err != nil {
		return stopFn, err
	}

	err = p.maybeStartCPUProfiling()
	if err != nil {
		return stopFn, err
	}

	err = p.maybeStartMemProfiling()
	if err != nil {
		return stopFn, err
	}

	return stopFn, nil
}

func (p *profiler) parseEnv() error {
	parser := envparser.NewParser()

	err := parser.Parse(&p.opts)
	if err != nil {
		return fmt.Errorf("cannot parse env vars: %w", err)
	}

	return nil
}

func (p *profiler) maybeStartCPUProfiling() error {
	if p.opts.OutPathCPU == "" {
		return nil
	}

	fcpu, err := os.Create(p.opts.OutPathCPU)
	if err != nil {
		return fmt.Errorf("cannot create cpu profile output: %w", err)
	}

	err = pprof.StartCPUProfile(fcpu)
	if err != nil {
		return fmt.Errorf("cannot start cpu profiling: %w", err)
	}

	p.addStopFn(func() {
		defer fcpu.Close()

		pprof.StopCPUProfile()
	})

	return nil
}

func (p *profiler) maybeStartMemProfiling() error {
	if p.opts.OutPathMem == "" {
		// If no memory profiling is requested, we must explicitly set [runtime.MemProfileRate] to `0`
		// to disable memory profiling because default [runtime.MemProfileRate] value is 512 KB.
		// Golang's linker will disable memory profiling if no code after "deadcode" elimination stage
		// links heap profiling functions from `pprof` module, but that is not the case here.
		// Since we call [pprof.WriteHeapProfile] conditionally during execution, the compiler cannot
		// eliminate this call, and the linker cannot disable memory profiling.
		// Therefore, [runtime.MemProfileRate] has its non-zero default value, and this trigger memory profiling.
		// Hence, we must reset [runtime.MemProfileRate] to `0` to disable memory profiling.
		runtime.MemProfileRate = 0

		return nil
	}

	if p.opts.MemRate == 0 {
		return ErrProfilerMemRateNotPositive
	}

	fmem, err := os.Create(p.opts.OutPathMem)
	if err != nil {
		return fmt.Errorf("cannot create mem profile output: %w", err)
	}

	runtime.MemProfileRate = int(p.opts.MemRate)

	p.addStopFn(func() {
		defer fmem.Close()

		err := pprof.WriteHeapProfile(fmem)
		if err != nil {
			p.logger.Error().WithErr(err).Str("output", p.opts.OutPathMem).Msg("cannot write mem profile")
		}
	})

	return nil
}

func (p *profiler) stopProfiling() {
	for _, fn := range p.stopFns {
		fn()
	}
}

func (p *profiler) addStopFn(fn profilerStopFunc) {
	p.stopFns = append(p.stopFns, fn)
}
