package tracking

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
	"golang.org/x/sync/semaphore"
)

// Processor processes hits
type Processor interface {
	ProcessHit(hit HitInterface) bool
}

// BatchHitProcessor batches hit into a queue and sends batch hit to the Data collect regularly or when queue is full
type BatchHitProcessor struct {
	envID         string
	apiKey        string
	MaxQueueSize  int           // max size of the queue before flush
	FlushInterval time.Duration // in milliseconds
	BatchSize     int
	Q             Queue
	flushLock     sync.Mutex
	Ticker        *time.Ticker
	HitDispatcher Dispatcher
	processing    *semaphore.Weighted
}

// DefaultBatchSize holds the default value for the batch size
const DefaultBatchSize = 10

// DefaultHitQueueSize holds the default value for the hit queue size
const DefaultHitQueueSize = 2000

// DefaultHitFlushInterval holds the default value for the hit flush interval
const DefaultHitFlushInterval = 30 * time.Second

const maxFlushWorkers = 1

var pLogger = logging.GetLogger("HitProcessor")

// BPOptionConfig is the BatchProcessor options that give you the ability to add one more more options before the processor is initialized.
type BPOptionConfig func(qp *BatchHitProcessor)

// WithBatchSize sets the batch size as a config option to be passed into the NewProcessor method
func WithBatchSize(bsize int) BPOptionConfig {
	return func(qp *BatchHitProcessor) {
		qp.BatchSize = bsize
	}
}

// WithQueueSize sets the queue size as a config option to be passed into the NewProcessor method
func WithQueueSize(qsize int) BPOptionConfig {
	return func(qp *BatchHitProcessor) {
		qp.MaxQueueSize = qsize
	}
}

// WithFlushInterval sets the flush interval as a config option to be passed into the NewProcessor method
func WithFlushInterval(flushInterval time.Duration) BPOptionConfig {
	return func(qp *BatchHitProcessor) {
		qp.FlushInterval = flushInterval
	}
}

// WithAPIKey sets the SDKKey used to register for notifications.  This should be removed when the project
// config supports sdk key.
func WithAPIKey(apiKey string) BPOptionConfig {
	return func(qp *BatchHitProcessor) {
		qp.apiKey = apiKey
	}
}

// NewBatchHitProcessor returns a new instance of BatchHitProcessor with queueSize and flushInterval
func NewBatchHitProcessor(envID string, options ...BPOptionConfig) *BatchHitProcessor {
	p := &BatchHitProcessor{
		envID:      envID,
		processing: semaphore.NewWeighted(int64(maxFlushWorkers)),
	}

	for _, opt := range options {
		opt(p)
	}

	if p.MaxQueueSize == 0 {
		p.MaxQueueSize = defaultQueueSize
	}

	if p.FlushInterval == 0 {
		p.FlushInterval = DefaultHitFlushInterval
	}

	if p.BatchSize == 0 {
		p.BatchSize = DefaultBatchSize
	}

	if p.BatchSize > p.MaxQueueSize {
		pLogger.Warning(
			fmt.Sprintf("Batch size %d is larger than queue size %d.  Setting to defaults",
				p.BatchSize, p.MaxQueueSize))

		p.BatchSize = DefaultBatchSize
		p.MaxQueueSize = defaultQueueSize
	}

	if p.Q == nil {
		p.Q = NewInMemoryQueue(p.MaxQueueSize)
	}

	if p.HitDispatcher == nil {
		dispatcher := NewQueueHitDispatcher(NewAPIClient(envID, DecisionAPIKey(p.apiKey)))
		p.HitDispatcher = dispatcher
	}

	return p
}

// Start does not do any initialization, just starts the ticker
func (p *BatchHitProcessor) Start(ctx context.Context) {
	pLogger.Info("Batch hit processor started")
	p.startTicker(ctx)
}

// ProcessHit takes the given user hit (can be an impression or conversion hit) and queues it up to be dispatched
// to the datacollect endpoint
func (p *BatchHitProcessor) ProcessHit(visitorID string, hit HitInterface) bool {
	if p.Q.Size() >= p.MaxQueueSize {
		pLogger.Warning("MaxQueueSize has been met. Discarding hit")
		return false
	}

	hit.setBaseInfos(p.envID, visitorID)
	errs := hit.validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Hit validation error : %v", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return false
	}

	p.Q.Add(hit)

	if p.Q.Size() < p.BatchSize {
		return true
	}

	if p.processing.TryAcquire(1) {
		// it doesn't matter if the timer has kicked in here.
		// we just want to start one go routine when the batch size is met.
		pLogger.Info("batch size reached.  Flushing routine being called")
		go func() {
			p.flushHits()
			p.processing.Release(1)
		}()
	}

	return true
}

// hitsCount returns size of an hit queue
func (p *BatchHitProcessor) hitsCount() int {
	return p.Q.Size()
}

// getHits returns hits from hit queue for count
func (p *BatchHitProcessor) getHits(count int) []HitInterface {
	return p.Q.Get(count)
}

// remove removes hits from queue for count
func (p *BatchHitProcessor) remove(count int) []HitInterface {
	return p.Q.Remove(count)
}

// StartTicker starts new ticker for flushing hits
func (p *BatchHitProcessor) startTicker(ctx context.Context) {
	if p.Ticker != nil {
		return
	}
	p.Ticker = time.NewTicker(p.FlushInterval)

	for {
		select {
		case <-p.Ticker.C:
			pLogger.Info("Hit processor ticked, flushing hits")
			p.flushHits()
		case <-ctx.Done():
			pLogger.Info("Hit processor stopped, flushing hits.")
			p.flushHits()
			d, ok := p.HitDispatcher.(*QueueHitDispatcher)
			if ok {
				d.flushHits()
			}
			return
		}
	}
}

// add the visitor to the current batch
func (p *BatchHitProcessor) addToBatch(current *BatchHit, hit HitInterface) {
	// Empty cid and vid for lighter payload
	hit.resetBaseHit()
	current.Hits = append(current.Hits, hit)
}

// flushHits flushes hits in queue
func (p *BatchHitProcessor) flushHits() {
	// we flush when queue size is reached.
	// however, if there is a ticker cycle already processing, we should wait
	p.flushLock.Lock()
	defer p.flushLock.Unlock()

	var batchHit BatchHit
	var batchHitCount = 0
	var failedToSend = false

	for p.hitsCount() > 0 {
		pLogger.Info("Handling hits")
		if failedToSend {
			pLogger.Error("last Hit Batch failed to send; retry on next flush", errors.New("dispatcher failed"))
			break
		}
		hits := p.getHits(p.BatchSize)

		if len(hits) > 0 {
			for i := 0; i < len(hits); i++ {
				hit := hits[i]
				if batchHitCount == 0 {
					batchHit = createBatchHit(hit)
				} else {
					p.addToBatch(&batchHit, hit)
				}
				batchHitCount++

				if batchHitCount >= p.BatchSize {
					// the batch size is reached so take the current batchHit and send it.
					break
				}
			}
		}

		if batchHitCount > 0 {
			toDispatch := batchHit
			if success, _ := p.HitDispatcher.DispatchHit(&toDispatch); success {
				pLogger.Debug("Dispatched event successfully")
				p.remove(batchHitCount)
				batchHitCount = 0
				batchHit = BatchHit{}
			} else {
				pLogger.Warning("Failed to dispatch event successfully")
				failedToSend = true
			}
		}
	}
}
