package loadgen

import "context"

// SchedulerHooks defines lifecycle events for a [Scheduler].
type SchedulerHooks interface {
	// Started is called when the scheduler loop has started.
	Started(ctx context.Context)

	// Executed is called after the [Executor.Execute] is called.
	Executed(ctx context.Context, err error)
}

// NoopSchedulerHooks does nothing.
type NoopSchedulerHooks struct{}

func (h NoopSchedulerHooks) Started(ctx context.Context) {}

func (h NoopSchedulerHooks) Executed(ctx context.Context, err error) {}
