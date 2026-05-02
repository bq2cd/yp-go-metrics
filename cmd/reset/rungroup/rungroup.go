// Package rungroup is a thin wrapper on top of `golang.org/x/sync/errgroup`, allowing to launch multiple objects
// conforming to `Run(context.Context) error` interface with a single function call.
// It also mandates a use of [context.Context] for cancellation (although objects can ignore it).
package rungroup

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Runnable defines an object that performs some long-running operation (typically, in a goroutine).
// It is expected that the object properly handles cancellation of the provided [context.Context]
// and terminates its execution as soon as possible.
type Runnable interface {
	Run(ctx context.Context) error
}

// Group is wrapper on top of [errgroup.Group] and provides a similar API but with a few differences.
// The main difference is [Group.Go] method, which accepts [context.Context] and multiple
// [Runnable] objects, then launches them in goroutines using underlying [errgroup.Group.Go] method.
type Group struct {
	egrp *errgroup.Group
}

// New creates an instance of [Group] with base [context.Context].
// A new context with cancellation is derived from the base context (using [errgroup.WithContext] method) and
// is returned alongside the group. This context is canceled when either a [Runnable] returns an error
// or all spawned goroutines finish their execution (causing [Group.Wait] to return), whichever happens first.
func New(baseCtx context.Context) (*Group, context.Context) {
	egrp, ctx := errgroup.WithContext(baseCtx)

	g := &Group{
		egrp: egrp,
	}

	return g, ctx
}

// Go launches provided [Runnable] objects in goroutines, passing to them
// provided context. The runnable object is expected to handle context's cancellation
// properly and terminate as fast as possible.
func (g *Group) Go(ctx context.Context, runnables ...Runnable) {
	for _, runnable := range runnables {
		g.egrp.Go(func() error {
			return runnable.Run(ctx)
		})
	}
}

// Wait simply waits for all spawned goroutines to finish their work and
// returns either the first error from any finished goroutine, or `nil` if
// all goroutines finished successfully.
// Wait relies on [errgroup.Group.Wait] under the hood.
func (g *Group) Wait() error {
	return g.egrp.Wait()
}
