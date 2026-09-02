package storage

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// FlushHandler defines the callback for batch writing to storage
type FlushHandler func(ctx context.Context, usages []domain.UsageEvent, costs []domain.CostItem) error

// MicroBatcher collects events and flushes them by batch size or time interval
type MicroBatcher struct {
	maxBatchSize  int
	flushInterval time.Duration
	handler       FlushHandler

	usageChan chan domain.UsageEvent
	costChan  chan domain.CostItem

	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewMicroBatcher(maxBatchSize int, flushIntervalMs int, handler FlushHandler) *MicroBatcher {
	if maxBatchSize <= 0 {
		maxBatchSize = 500
	}
	if flushIntervalMs <= 0 {
		flushIntervalMs = 200
	}

	b := &MicroBatcher{
		maxBatchSize:  maxBatchSize,
		flushInterval: time.Duration(flushIntervalMs) * time.Millisecond,
		handler:       handler,
		usageChan:     make(chan domain.UsageEvent, maxBatchSize*10),
		costChan:      make(chan domain.CostItem, maxBatchSize*10),
		stopChan:      make(chan struct{}),
	}

	b.wg.Add(1)
	go b.runLoop()

	return b
}

// Push adds a usage event and its corresponding cost item to the buffer
func (b *MicroBatcher) Push(usage domain.UsageEvent, cost domain.CostItem) {
	select {
	case b.usageChan <- usage:
	default:
		log.Printf("[WARN] Usage channel full, dropping or backpressuring")
	}

	select {
	case b.costChan <- cost:
	default:
		log.Printf("[WARN] Cost channel full, dropping or backpressuring")
	}
}

func (b *MicroBatcher) runLoop() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	usageBatch := make([]domain.UsageEvent, 0, b.maxBatchSize)
	costBatch := make([]domain.CostItem, 0, b.maxBatchSize)

	flush := func() {
		if len(usageBatch) == 0 && len(costBatch) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := b.handler(ctx, usageBatch, costBatch); err != nil {
			log.Printf("[ERROR] Failed to flush batch: %v", err)
		}

		usageBatch = make([]domain.UsageEvent, 0, b.maxBatchSize)
		costBatch = make([]domain.CostItem, 0, b.maxBatchSize)
	}

	for {
		select {
		case <-b.stopChan:
			// Drain remaining events
			for {
				select {
				case u := <-b.usageChan:
					usageBatch = append(usageBatch, u)
				default:
					goto drainedUsage
				}
			}
		drainedUsage:
			for {
				select {
				case c := <-b.costChan:
					costBatch = append(costBatch, c)
				default:
					goto drainedCost
				}
			}
		drainedCost:
			flush()
			return

		case u := <-b.usageChan:
			usageBatch = append(usageBatch, u)
			if len(usageBatch) >= b.maxBatchSize {
				flush()
			}

		case c := <-b.costChan:
			costBatch = append(costBatch, c)
			if len(costBatch) >= b.maxBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

// Stop gracefully stops the batcher and flushes any buffered items
func (b *MicroBatcher) Stop() {
	close(b.stopChan)
	b.wg.Wait()
}
