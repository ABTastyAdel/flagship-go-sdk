package tracking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createHit() *EventHit {
	hit := &EventHit{
		Action: "action",
	}
	hit.setBaseInfos(testEnvID, "test_vid")
	return hit
}

func TestQueueHitDispatcher_DispatchHit(t *testing.T) {
	q := NewQueueHitDispatcher(NewMockAPIClient(testEnvID, false))
	batch := createBatchHit(createHit())

	success, _ := q.DispatchHit(&batch)

	assert.True(t, success)

	// its been queued
	assert.True(t, (q.hitQueue.Size() == 1) || (q.hitQueue.Size() == 0))

	// give the queue a chance to run
	time.Sleep(1 * time.Second)

	// check the queue
	assert.Equal(t, 0, q.hitQueue.Size())
}

func TestQueueHitDispatcher_InvalidHit(t *testing.T) {
	q := NewQueueHitDispatcher(NewMockAPIClient(testEnvID, false))
	batch := createBatchHit(createHit())

	q.hitQueue.Add(&batch)

	assert.Equal(t, 1, q.hitQueue.Size())

	// give the queue a chance to run
	q.flushHits()

	// check the queue. bad event type should be removed.  but, not sent.
	assert.Equal(t, 0, q.hitQueue.Size())

}

func TestQueueHitDispatcher_FailDispath(t *testing.T) {
	q := NewQueueHitDispatcher(NewMockAPIClient(testEnvID, true))

	event := createHit()
	batchHit := createBatchHit(event)
	q.DispatchHit(&batchHit)

	assert.Equal(t, 1, q.hitQueue.Size())

	// Flush the hits
	q.flushHits()
	time.Sleep(1 * time.Second)

	// Error on api call should not remove the item from the queue
	assert.Equal(t, 1, q.hitQueue.Size())

	q.flushHits()

	// Error on api call should not remove the item from the queue
	assert.Equal(t, 1, q.hitQueue.Size())
}
