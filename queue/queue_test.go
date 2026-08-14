package queue

import (
	"testing"
)

func TestProcessor(t *testing.T) {
	q := NewProcessor(2)

	// Try acquire 1 - success
	if err := q.TryAcquire(false); err != nil {
		t.Errorf("Expected success, got %v", err)
	}

	// Try acquire 2 - success
	if err := q.TryAcquire(false); err != nil {
		t.Errorf("Expected success, got %v", err)
	}

	// Try acquire 3 - should fail
	if err := q.TryAcquire(false); err != ErrQueueFull {
		t.Errorf("Expected ErrQueueFull, got %v", err)
	}

	// Try acquire 4 (Owner) - should succeed despite being full
	if err := q.TryAcquire(true); err != nil {
		t.Errorf("Expected success for owner, got %v", err)
	}

	q.Release()
	q.Release()

	// Now should have 1 free slot for non-owner
	if err := q.TryAcquire(false); err != nil {
		t.Errorf("Expected success after release, got %v", err)
	}
}
