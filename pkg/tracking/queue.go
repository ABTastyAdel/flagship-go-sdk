package tracking

import (
	"sync"
)

// Queue represents a queue
type Queue interface {
	Add(item HitInterface)
	Remove(count int) []HitInterface
	Get(count int) []HitInterface
	Size() int
}

// InMemoryQueue represents a in-memory queue
type InMemoryQueue struct {
	Queue []HitInterface
	Mux   sync.Mutex
}

// Get returns queue for given count size
func (i *InMemoryQueue) Get(count int) []HitInterface {
	if i.Size() < count {
		count = i.Size()
	}
	i.Mux.Lock()
	defer i.Mux.Unlock()
	return i.Queue[:count]
}

// Add appends item to queue
func (i *InMemoryQueue) Add(item HitInterface) {
	i.Mux.Lock()
	i.Queue = append(i.Queue, item)
	i.Mux.Unlock()
}

// Remove removes item from queue by count and returns elements slice
func (i *InMemoryQueue) Remove(count int) []HitInterface {
	if i.Size() < count {
		count = i.Size()
	}
	i.Mux.Lock()
	defer i.Mux.Unlock()
	elem := i.Queue[:count]
	i.Queue = i.Queue[count:]
	return elem
}

// Size returns size of queue
func (i *InMemoryQueue) Size() int {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	return len(i.Queue)
}

// NewInMemoryQueue returns new InMemoryQueue with given queueSize
func NewInMemoryQueue(queueSize int) Queue {
	i := &InMemoryQueue{Queue: make([]HitInterface, 0, queueSize)}
	return i
}
