package pipeline

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/sync/errgroup"
)

// InitCloser describes an object that can be initialized and closed.
type InitCloser interface {
	Init(ctx context.Context) error
	Close(ctx context.Context) error
}

// runStage is responsible for pipeline's stage lifecycle:
// it initializes stage's processor, launches worker pool to process the data, and
// finally closes the processor.
func runStage(ctx context.Context, name string, initCloser InitCloser, runFn func(context.Context) error) error {
	err := initCloser.Init(ctx)
	if err != nil {
		return fmt.Errorf("cannot init %s: %w", name, err)
	}

	errFinal := runFn(ctx)

	err = initCloser.Close(ctx)
	if err != nil {
		errFinal = errors.Join(errFinal, fmt.Errorf("cannot close %s: %w", name, err))
	}

	return errFinal
}

// runWorkers launches provided `workerFn` in parallel, using [errgroup.Group] with limit, then
// waits for their completion. If `count` is zero, a worker per CPU ([runtime.NumCPU]) will be launched.
// After all workers are finished, a `cleanupFn` is executed (if not `nil`).
func runWorkers(
	baseCtx context.Context,
	count uint,
	workerFn func(context.Context) error,
	cleanupFn func(),
) error {
	defer func() {
		if cleanupFn != nil {
			cleanupFn()
		}
	}()

	if count == 0 { // ensure we actually run workers (worker per CPU is fine in this case)
		count = uint(runtime.NumCPU())
	}

	grp, ctx := errgroup.WithContext(baseCtx)
	for range count {
		grp.Go(func() error {
			return workerFn(ctx)
		})
	}

	return grp.Wait()
}

// runProcessingLoop creates an infinite loop that consumes an item from the incoming channel, calls
// provided `processorFn` on the obtained item, then optionally send the result to the outgoing channel (if not `nil`).
// The loop aborts either when `processorFn` returns an error, or when provided context is canceled.
// It is expected that `processorFn` terminates on context cancellation as soon as possible.
func runProcessingLoop[I, O any](
	ctx context.Context,
	name string,
	inCh <-chan I,
	outCh chan<- O,
	processorFn func(context.Context, I) (O, error),
) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s aborted: %w", name, ctx.Err())
		case item, ok := <-inCh:
			if !ok { // channel is closed and no data is left in it
				return nil
			}

			result, err := processorFn(ctx, item)
			if err != nil {
				return fmt.Errorf("%s cannot process item %v: %w", name, item, err)
			}

			if outCh == nil { // intentionally discard results
				continue loop
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("%s aborted: %w", name, ctx.Err())
			case outCh <- result:
				continue loop
			}
		}
	}
}
