package island

import (
	"sync"
	"testing"
	"time"
)

type syncerStub struct {
	mu      sync.Mutex
	called  int
	updates []Update
	done    chan struct{}
}

func (s *syncerStub) SyncUpdates(updates []Update) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.called++
	s.updates = append([]Update(nil), updates...)
	if s.done != nil {
		close(s.done)
		s.done = nil
	}
	return nil
}

func TestNewManagerInitialMode(t *testing.T) {
	online := true
	mgr := NewManager(5*time.Millisecond, 10, func() bool { return online })

	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
	if mgr.GetMode() != ModeOnline {
		t.Fatalf("expected initial mode online, got %v", mgr.GetMode())
	}
}

func TestModeTransitionAndSync(t *testing.T) {
	online := true
	mgr := NewManager(5*time.Millisecond, 5, func() bool { return online })
	stub := &syncerStub{done: make(chan struct{})}
	mgr.SetSyncer(stub)

	if err := mgr.CacheUpdate(Update{Round: 1, Timestamp: time.Now(), PeerID: "n1"}); err != nil {
		t.Fatalf("cache update failed: %v", err)
	}

	online = false
	mgr.updateMode(false)
	if mgr.GetMode() != ModeIsland {
		t.Fatalf("expected island mode, got %v", mgr.GetMode())
	}

	online = true
	mgr.updateMode(true)

	select {
	case <-stub.done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected cached updates to be synced when returning online")
	}

	if mgr.GetMode() != ModeOnline {
		t.Fatalf("expected online mode, got %v", mgr.GetMode())
	}

	cached, _ := mgr.GetCachedUpdateStats()
	if cached != 0 {
		t.Fatalf("expected cache to be empty after sync, got %d", cached)
	}
}

func TestCacheEvictsOldestWhenFull(t *testing.T) {
	mgr := NewManager(10*time.Millisecond, 2, func() bool { return true })

	_ = mgr.CacheUpdate(Update{Round: 1, Timestamp: time.Now(), PeerID: "a"})
	_ = mgr.CacheUpdate(Update{Round: 2, Timestamp: time.Now(), PeerID: "b"})
	_ = mgr.CacheUpdate(Update{Round: 3, Timestamp: time.Now(), PeerID: "c"})

	updates := mgr.GetCachedUpdates()
	if len(updates) != 2 {
		t.Fatalf("expected 2 cached updates, got %d", len(updates))
	}
	if updates[0].Round != 2 || updates[1].Round != 3 {
		t.Fatalf("expected oldest update to be evicted, got rounds %d and %d", updates[0].Round, updates[1].Round)
	}
}
