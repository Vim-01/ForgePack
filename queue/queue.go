package queue

import (
	"errors"
	"sync"
)

var ErrQueueFull = errors.New("очередь переполнена, попробуйте позже")

type Processor struct {
	maxWorkers int
	active     int
	mu         sync.Mutex
}

func NewProcessor(max int) *Processor {
	return &Processor{
		maxWorkers: max,
	}
}

// TryAcquire returns true if a worker slot was acquired
func (p *Processor) TryAcquire(isOwner bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Owner can always bypass the limit
	if isOwner {
		p.active++
		return nil
	}

	if p.active >= p.maxWorkers {
		return ErrQueueFull
	}

	p.active++
	return nil
}

func (p *Processor) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active > 0 {
		p.active--
	}
}
