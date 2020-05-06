package tracking

import (
	"fmt"
	"sync"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

const maxRetries = 3
const defaultQueueSize = 1000
const sleepTime = 1 * time.Second

var dispatcherLogger = logging.GetLogger("HitDispatcher")

// Dispatcher dispatches hits
type Dispatcher interface {
	DispatchHit(hit HitInterface) (bool, error)
}

// BatchDispatcher dispatches batch hits
type BatchDispatcher interface {
	DispatchHit(hit *BatchHit) (bool, error)
}

// HTTPHitDispatcher is the HTTP implementation of the Dispatcher interface
type HTTPHitDispatcher struct {
	trackingAPIClient APIClientInterface
}

// DispatchHit dispatches hit with callback
func (ed *HTTPHitDispatcher) DispatchHit(hit *BatchHit) (bool, error) {
	dispatcherLogger.Info("Dispatching hit to collect")
	for _, hit := range hit.Hits {
		hit.computeQueueTime()
	}

	err := ed.trackingAPIClient.sendInternalHit(hit)
	var success bool
	if err != nil {
		dispatcherLogger.Error("Hit sending failed :", err)
		success = false
	}
	success = true
	return success, err
}

// QueueHitDispatcher is a queued version of the hit Dispatcher that queues, returns success, and dispatches hits in the background
type QueueHitDispatcher struct {
	hitQueue        Queue
	hitFlushLock    sync.Mutex
	BatchDispatcher BatchDispatcher
}

// DispatchHit queues hit with callback and calls flush in a go routine.
func (ed *QueueHitDispatcher) DispatchHit(hit HitInterface) (bool, error) {
	ed.hitQueue.Add(hit)
	go func() {
		ed.flushHits()
	}()
	return true, nil
}

// flush the hits
func (ed *QueueHitDispatcher) flushHits() {
	ed.hitFlushLock.Lock()

	defer func() {
		ed.hitFlushLock.Unlock()
	}()

	retryCount := 0
	for ed.hitQueue.Size() > 0 {
		if retryCount > maxRetries {
			dispatcherLogger.Error(fmt.Sprintf("hit failed to send %d times. It will retry on next hit sent", maxRetries), nil)
			break
		}

		items := ed.hitQueue.Get(1)
		if len(items) == 0 {
			// something happened.  Just continue and you should expect size to be zero.
			continue
		}
		hit := items[0]

		batchHit := hit.(*BatchHit)

		success, err := ed.BatchDispatcher.DispatchHit(batchHit)

		if err == nil {
			if success {
				dispatcherLogger.Debug(fmt.Sprintf("Dispatched log hit %+v", hit))
				ed.hitQueue.Remove(1)
				retryCount = 0
			} else {
				dispatcherLogger.Warning("dispatch hit failed")
				// Dispatch hit failed.  Sleep some seconds and try again.
				time.Sleep(sleepTime)
				retryCount++
			}
		} else {
			dispatcherLogger.Error("Error dispatching ", err)
			// we failed.  Sleep some seconds and try again.
			time.Sleep(sleepTime)
			// increase retryCount.  We exit if we have retried x times.
			// we will retry again next hit that is added.
			retryCount++
		}
	}
}

// NewQueueHitDispatcher creates a Dispatcher that queues in memory and then sends via go routine.
func NewQueueHitDispatcher(trackingAPIClient APIClientInterface) *QueueHitDispatcher {
	return &QueueHitDispatcher{
		hitQueue:        NewInMemoryQueue(defaultQueueSize),
		BatchDispatcher: &HTTPHitDispatcher{trackingAPIClient: trackingAPIClient},
	}
}
