// Package sharedchan provides a wrapper on top of built-in channel primitive to allow multiple goroutines
// to send to the wrapped channel without being responsible for closing it.
// The wrapped channel can be safely closed from a totally separate goroutine (e.g. a monitoring goroutine),
// and senders will be "notified" when this happens.
package sharedchan

import (
	"sync"
	"sync/atomic"
)

// Channel wraps a built-in channel primitive of type `T` so that it can be safely closed
// from a goroutine different from (multiple) sender goroutine(s).
// Sender goroutines will be notified that the channel is closed by return value
// of [Channel.Send] method.
type Channel[T any] struct {
	incoming chan T // points to the same channel
	outgoing chan T // points to the same channel
	closing  chan struct{}
	closed   atomic.Bool
	mu       sync.RWMutex
}

// NewChannel creates an instance of [Channel] with underlying channel and buffer of given `size`.
// If provided `size` is zero, the underlying channel is unbuffered.
func NewChannel[T any](size uint) *Channel[T] {
	ch := make(chan T, size)

	return &Channel[T]{
		incoming: ch,
		outgoing: ch,
		closing:  make(chan struct{}),
	}
}

// Send will attempt to send provided value to the underlying channel and will return `true`
// if the value was sent successfully (accepted by the channel); otherwise, `false` will be returned.
// Send will block if the channel is full, waiting either for someone to consume from the channel
// (and thus make room for sending) or for the channel's closure.
func (c *Channel[T]) Send(value T) bool {
	// although we're writing into `c.incoming` channel, we access `c.incoming` variable
	// in "read" mode (that is, we're "reading" its value); therefore, we need a read
	// lock here.
	c.mu.RLock()
	defer c.mu.RUnlock()

	select {
	case <-c.closing:
		return false
	case c.incoming <- value:
		return true
	}
}

// Receive returns underlying channel in reading mode.
// This will the callers to use `select` or other standard means to receive
// from the channel and get notified about the channel's closure.
func (c *Channel[T]) Receive() <-chan T {
	return c.outgoing
}

// Close safely closes the underlying channel, ensuring that the channel is closed
// only once and preventing senders from sending to it by coordinating with [Channel.Send] method.
// Close is safe to be called concurrently and multiple times - only the first call will
// actually close the channel, the others will return immediately.
func (c *Channel[T]) Close() {
	// ensure only the first call actually performs closing.
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	// send signal to all blocked senders that are waiting on `c.incoming` to become ready;
	// this will unblock them and force to release their read locks.
	close(c.closing)

	// taking write lock to modify `c.incoming` variable; this will wait until all
	// "readers" in [Channel.Send] method release their read locks.
	c.mu.Lock()
	// `select` will ignore `nil` channels by design but closed channels are always ready
	// (both for sending and receiving); once set to `nil`, [Channel.Send] won't pick this channel anymore.
	c.incoming = nil
	c.mu.Unlock()

	// now we can safely close the channel (`c.outgoing` and `c.incoming` point to the same channel):
	// since `c.incoming` is `nil`, the senders won't attempt to send to it anymore.
	close(c.outgoing)
}
