package main

import (
	"context"
	"fmt"
	"time"
)

// defaultInterval used for await.
const defaultInterval = time.Second

// Await the completion of an operation. The operation is considered complete
// when f() returns true. If f() returns and error, await exits early and returns
// an error.
//
// If interval is zero, it defaults to defaultInterval.
//
// To implement a timeout, use context.WithDeadline():
//   ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute))
//   defer cancel()
//   await(ctx, ...)
func await(
	ctx context.Context,
	message string,
	f func(context.Context) (bool, error),
	interval time.Duration,
) error {
	if interval == 0 {
		interval = defaultInterval
	}
	if !quiet {
		fmt.Print(message)
		defer fmt.Println(" Done!")
	}
	delay := time.NewTimer(0) // No delay on first attempt.
	defer delay.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())

		case <-delay.C:
			if !quiet {
				fmt.Print(".")
			}
			ok, err := f(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			delay.Reset(interval)
		}
	}
}
