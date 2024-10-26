package loadgen

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

// Scheduler schedules calls to an [Executor].
//
// See [NewScheduler].
type Scheduler struct {
	wg         sync.WaitGroup
	executor   Executor
	hooks      SchedulerHooks
	qps        int64
	maxWorkers int64
}

// SchedulerOption is a configurable option for [NewScheduler].
type SchedulerOption func(*Scheduler)

// WithHooks sets the [SchedulerHooks] for a [Scheduler].
func WithHooks(h SchedulerHooks) SchedulerOption {
	return func(s *Scheduler) {
		s.hooks = h
	}
}

// NewScheduler returns a new [Scheduler].
func NewScheduler(
	executor Executor,
	qps int64,
	maxWorkers int64,
	opts ...SchedulerOption,
) *Scheduler {
	s := &Scheduler{
		executor:   executor,
		qps:        qps,
		maxWorkers: maxWorkers,
		hooks:      NoopSchedulerHooks{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run starts the scheduler loop and blocks until the ctx is done. It returns
// ctx.Err.
func (s *Scheduler) Run(ctx context.Context) error {
	limiter := rate.NewLimiter(rate.Limit(s.qps), 1)
	sem := semaphore.NewWeighted(s.maxWorkers)

	s.hooks.Started(ctx)

	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return ctx.Err()
		default:
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("acquire semaphore: %v", err)
		}

		if err := limiter.Wait(ctx); err != nil {
			return fmt.Errorf("limiter wait: %v", err)
		}

		s.wg.Add(1)
		go func() {
			defer sem.Release(1)
			defer s.wg.Done()

			err := s.executor.Execute(ctx)

			s.hooks.Executed(ctx, err)
		}()
	}
}
