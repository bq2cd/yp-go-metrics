package pipeline

const (
	// DefaultConfigParallelWorkers is the number of workers per stage to launch in parallel.
	DefaultConfigParallelWorkers = 2
	// DefaultConfigOutputBufferSize is the size of the output channel for [Step] and [Source] stages.
	DefaultConfigOutputBufferSize = 2
)

// Config contains commons configuration knobs for all stages.
type Config struct {
	ParallelWorkers  uint
	OutputBufferSize uint
}

// DefaultConfig produces [Config] with default values.
func DefaultConfig() Config {
	return Config{
		ParallelWorkers:  DefaultConfigParallelWorkers,
		OutputBufferSize: DefaultConfigOutputBufferSize,
	}
}
