package presence

import (
	"context"
	"log"
	"sync"
	"time"
	"users/internal/broker"
	"users/internal/repository"

	"github.com/google/uuid"
)

const (
	HeartbeatTimeout = 25 * time.Second
	ReaperInterval   = 10 * time.Second
)

// Tracker maintains in-memory presence state and drives online/offline transitions.
type Tracker struct {
	mu     sync.Mutex
	beats  map[uuid.UUID]time.Time // userId → last heartbeat wall-clock time
	online map[uuid.UUID]bool      // current known state
	broker *broker.Broker
	repo   repository.IUsersRepository
}

func NewTracker(b *broker.Broker, repo repository.IUsersRepository) *Tracker {
	return &Tracker{
		beats:  make(map[uuid.UUID]time.Time),
		online: make(map[uuid.UUID]bool),
		broker: b,
		repo:   repo,
	}
}

// Heartbeat records the latest heartbeat for userID.
// On the first heartbeat after being offline it writes last_seen_at to the DB
// and broadcasts an online event to all SSE subscribers.
func (t *Tracker) Heartbeat(ctx context.Context, userID uuid.UUID) {
	t.mu.Lock()
	wasOnline := t.online[userID]
	t.beats[userID] = time.Now()
	t.online[userID] = true
	t.mu.Unlock()

	if !wasOnline {
		if err := t.repo.TouchLastSeen(ctx, userID); err != nil {
			log.Printf("presence: touch last_seen_at for %s: %v", userID, err)
		}
		t.broker.Publish("global", broker.PresenceEvent{
			UserID:   userID.String(),
			Online:   true,
			LastSeen: time.Now().Unix(),
		})
		log.Printf("presence: %s is now online", userID)
	}
}

// OnlineUsers returns a snapshot of currently-online user IDs.
// Used to seed newly-connected SSE clients with the current state.
func (t *Tracker) OnlineUsers() map[string]bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[string]bool, len(t.online))
	for id, on := range t.online {
		if on {
			out[id.String()] = true
		}
	}
	return out
}

// StartReaper launches the background goroutine that marks stale users offline.
// It runs until ctx is cancelled.
func (t *Tracker) StartReaper(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(ReaperInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.reap(ctx)
			}
		}
	}()
}

func (t *Tracker) reap(ctx context.Context) {
	now := time.Now()
	t.mu.Lock()
	var stale []uuid.UUID
	for id, lastBeat := range t.beats {
		if t.online[id] && now.Sub(lastBeat) > HeartbeatTimeout {
			stale = append(stale, id)
			t.online[id] = false
		}
	}
	t.mu.Unlock()

	for _, id := range stale {
		if err := t.repo.TouchLastSeen(ctx, id); err != nil {
			log.Printf("presence: reaper touch %s: %v", id, err)
		}
		t.broker.Publish("global", broker.PresenceEvent{
			UserID:   id.String(),
			Online:   false,
			LastSeen: now.Unix(),
		})
		log.Printf("presence: %s went offline (no heartbeat for >%s)", id, HeartbeatTimeout)
	}
}
