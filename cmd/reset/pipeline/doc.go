// Package pipeline provides primitives for launching multiple stages concurrently (each stage is a separate goroutine)
// while maintaining sequential data flow. The stages are connected via channels: incoming and outgoing, with the
// exception for the first and last stages, which have only one channel (outgoing and incoming, respectively).
// The data is passed to the next stage as soon as it has been processed by a previous one.
// Each stage can be configured to process incoming data in parallel (with a fixed number of workers), with minimal
// parallelism (2 workers) enabled by default.
// NB: Only serial pipelines (no parallel stages) are current supported,
// e.g. [Source] -> [Step] -> [Step] ... -> [Sink].
package pipeline
