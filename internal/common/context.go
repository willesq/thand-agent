package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithInterrupt creates a context that is cancelled when an interrupt signal
// (SIGINT or SIGTERM) is received. Returns the context and a cleanup function
// that should be called when done (typically via defer).
//
// Example usage:
//
//	ctx, cleanup := common.WithInterrupt(context.Background())
//	defer cleanup()
//	result := someBlockingOperation(ctx)
func WithInterrupt(parent context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
			// Context was cancelled by other means, exit goroutine
		}
	}()

	cleanup := func() {
		signal.Stop(sigChan)
		cancel()
	}

	return ctx, cleanup
}

// NewInterruptChannel creates a channel that receives interrupt signals
// (SIGINT or SIGTERM). Returns the channel and a cleanup function that
// should be called when done.
//
// Use this when you need to handle the signal directly (e.g., for graceful shutdown).
//
// Example usage:
//
//	sigChan, cleanup := common.NewInterruptChannel()
//	defer cleanup()
//	sig := <-sigChan
//	server.Stop()
func NewInterruptChannel() (<-chan os.Signal, func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	cleanup := func() {
		signal.Stop(sigChan)
	}

	return sigChan, cleanup
}
