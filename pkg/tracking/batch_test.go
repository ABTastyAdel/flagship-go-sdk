package tracking

import (
	"context"
	"testing"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestBatch_DispatchEvent(t *testing.T) {
	processor := NewBatchHitProcessor(testEnvID)
	processor.HitDispatcher = NewQueueHitDispatcher(NewMockAPIClient(testEnvID, false))

	eg := utils.NewExecGroup(context.Background())
	eg.Go(processor.Start)

	event := createHit()

	processor.ProcessHit(event.VisitorID, event)

	assert.Equal(t, 1, processor.hitsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.hitsCount())
}

func TestBatch_Create(t *testing.T) {
	processor := NewBatchHitProcessor(testEnvID,
		WithQueueSize(10),
		WithBatchSize(50),
		WithFlushInterval(100),
		WithAPIKey("test_key"))

	// Test batch size not > queue size
	assert.Equal(t, 10, processor.BatchSize)

	processor = NewBatchHitProcessor(testEnvID,
		WithQueueSize(10),
		WithBatchSize(5),
		WithFlushInterval(100),
		WithAPIKey("test_key"))
	processor.HitDispatcher = NewQueueHitDispatcher(NewMockAPIClient(testEnvID, false))

	assert.Equal(t, 5, processor.BatchSize)
	assert.Equal(t, 10, processor.MaxQueueSize)
	assert.Equal(t, time.Duration(100), processor.FlushInterval)
	assert.Equal(t, "test_key", processor.apiKey)

	eg := utils.NewExecGroup(context.Background())
	eg.Go(processor.Start)

	event := createHit()

	processor.ProcessHit(event.VisitorID, event)

	assert.Equal(t, 1, processor.hitsCount())

	hit := processor.getHits(1)[0]
	hitEvent := hit.(*EventHit)

	assert.Equal(t, event.Action, hitEvent.Action)

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.hitsCount())
}

func TestBatch_InvalidHit(t *testing.T) {
	processor := NewBatchHitProcessor(testEnvID, WithQueueSize(10),
		WithFlushInterval(100))
	processor.HitDispatcher = NewQueueHitDispatcher(NewMockAPIClient(testEnvID, false))

	eg := utils.NewExecGroup(context.Background())
	eg.Go(processor.Start)

	wrongEvent := &EventHit{}

	res, _ := processor.ProcessHit(testVisitorID, wrongEvent)

	assert.Equal(t, false, res)
	assert.Equal(t, 0, processor.hitsCount())

	eg.TerminateAndWait()
}

func TestBatch_MaxQueue(t *testing.T) {
	processor := NewBatchHitProcessor(testEnvID, WithQueueSize(10),
		WithFlushInterval(10*time.Second))
	processor.HitDispatcher = NewQueueHitDispatcher(NewMockAPIClient(testEnvID, false))

	eg := utils.NewExecGroup(context.Background())
	eg.Go(processor.Start)

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		event := createHit()
		processor.ProcessHit(testVisitorID, event)
	}

	assert.Equal(t, 10, processor.hitsCount())
	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, processor.hitsCount())
}
