package tracking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryQueue_Add_Size_Remove(t *testing.T) {
	q := NewInMemoryQueue(5)

	ev1 := &EventHit{}
	ev2 := &EventHit{}
	ev3 := &EventHit{}

	q.Add(ev1)
	q.Add(ev2)
	q.Add(ev3)

	assert.Equal(t, 3, q.Size())

	items1 := q.Get(1)

	assert.Equal(t, 1, len(items1))
	assert.Equal(t, ev1, items1[0])

	items2 := q.Get(5)

	assert.Equal(t, 3, len(items2))
	assert.Equal(t, ev3, items2[2])

	empty := q.Get(0)
	assert.Equal(t, 0, len(empty))

	allItems := q.Remove(3)

	assert.Equal(t, 3, len(allItems))

	assert.Equal(t, 0, q.Size())
}

func TestInMemoryQueue_Concurrent(t *testing.T) {

	q := NewInMemoryQueue(5)

	quit := make(chan int)

	go func() {
		i := 5
		for i > 0 {
			q.Add(&EventHit{})
			i--
		}

		quit <- 0
	}()

	go func() {
		i := 5
		for i > 0 {
			q.Add(&EventHit{})
			i--
		}

		quit <- 0
	}()

	<-quit

	q.Remove(1)
	q.Remove(1)

	<-quit

	assert.Equal(t, 8, q.Size())
}
