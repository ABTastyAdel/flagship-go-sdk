package utils

import (
	"context"
	"sync"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

var logger = logging.GetLogger("ExecGroup")

// ExecGroup is a utility for managing graceful, blocking cancellation of goroutines.
type ExecGroup struct {
	wg         *sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewExecGroup returns constructed object
func NewExecGroup(ctx context.Context) *ExecGroup {
	nctx, cancelFn := context.WithCancel(ctx)
	wg := sync.WaitGroup{}

	return &ExecGroup{wg: &wg, ctx: nctx, cancelFunc: cancelFn}
}

// Go initiates a goroutine with the inputted function. Each invocation increments a shared WaitGroup
// before being initiated. Once the supplied function exits the WaitGroup is decremented.
// A common ctx is passed to each input function to signal a shutdown sequence.
func (c ExecGroup) Go(f func(ctx context.Context)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		f(c.ctx)
	}()
}

// TerminateAndWait sends termination signal and waits
func (c ExecGroup) TerminateAndWait() {
	if c.cancelFunc == nil {
		logger.Error("failed to shut down Execution Context properly", nil)
		return
	}
	c.cancelFunc()
	c.wg.Wait()
}
