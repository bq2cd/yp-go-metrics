package cli

import "flag"

type profilerStopFunc func()

type profilerOpts struct {
}

type profiler struct {
	opts profilerOpts
}

func newProfiler() *profiler {
	return &profiler{
		opts: profilerOpts{},
	}
}

func (p *profiler) AddProfilingArgs(fs *flag.FlagSet) {

}

func (p *profiler) MaybeStartProfiling() (profilerStopFunc, error) {
	return func() {}, nil
}
