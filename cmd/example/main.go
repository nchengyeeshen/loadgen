package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/nchengyeeshen/loadgen"
)

var (
	qps        int64
	maxWorkers int64
	called     atomic.Int64
)

func main() { run(context.Background(), os.Args, os.Stderr) }

func run(ctx context.Context, args []string, stderr io.Writer) {
	// Parse command line flags.
	fs := flag.NewFlagSet("example", flag.PanicOnError)
	fs.Int64Var(&qps, "qps", 1, "desired QPS")
	fs.Int64Var(&maxWorkers, "max-workers", 1, "desired maximum number of worker goroutines")
	_ = fs.Parse(args[1:])

	// Create a new logger.
	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	s := loadgen.NewScheduler(
		myExecutor{},
		qps,
		maxWorkers,
		// Specify our custom hooks.
		loadgen.WithHooks(myHooks{logger: logger}),
	)

	// Create a new context for cancellation.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Listen for cancellation signals.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan
		logger.Debug("cancellation signal received, shutting down scheduler")
		cancel()
	}()

	if err := s.Run(ctx); err != nil {
		logger.WarnContext(ctx, "scheduler.run", "err", err)
	}

	logger.Debug("terminated")
}

// myExecutor increments a centralised counter when called.
type myExecutor struct {
}

func (f myExecutor) Execute(ctx context.Context) error {
	called.Add(1)
	return nil
}

type myHooks struct {
	logger *slog.Logger
}

func (h myHooks) Started(ctx context.Context) {
	h.logger.DebugContext(ctx, "scheduler started!")
}

func (h myHooks) Executed(ctx context.Context, err error) {
	h.logger.DebugContext(ctx, "executed", "times", called.Load())
}
