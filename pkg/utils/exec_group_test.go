package utils

import (
	"context"
	"sync"
	"testing"
)

func TestWithContextCancelFunc(t *testing.T) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	eg := NewExecGroup(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	cancelFunc()
	wg.Wait()
}

func TestTerminateAndWait(t *testing.T) {

	eg := NewExecGroup(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	eg.TerminateAndWait()
	wg.Wait()
}
