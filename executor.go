package loadgen

import "context"

// Executor defines an action that can be executed.
type Executor interface {
	// Execute an action.
	Execute(ctx context.Context) error
}
