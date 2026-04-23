// Package memstats defines a metric source to read agent process's runtime metrics
// using Go's [runtime.MemStats] interface.
// The collected metrics include various memory allocation counters, garbage collector stats
// and the number of running goroutines.
package memstats
