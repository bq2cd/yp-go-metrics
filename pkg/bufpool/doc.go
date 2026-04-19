// Package bufpool creates a thin wrapper on top of [sync.Pool] to encapsulate release of [bytes.Buffer]
// objects to the pool into [Close] method.
package bufpool
