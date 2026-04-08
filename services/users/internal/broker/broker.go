package broker

import "sync"

// PresenceEvent is broadcast to all presence stream subscribers.
type PresenceEvent struct {
	UserID   string `json:"user_id"`
	Online   bool   `json:"online"`
	LastSeen int64  `json:"last_seen"` // Unix seconds
}

// Broker is an in-memory pub/sub hub for presence SSE events. Safe for concurrent use.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan PresenceEvent
}

func NewBroker() *Broker {
	return &Broker{subscribers: make(map[string][]chan PresenceEvent)}
}

// Subscribe registers a new subscriber for key.
// The caller MUST call Unsubscribe(key, ch) when done (on client disconnect).
func (b *Broker) Subscribe(key string) chan PresenceEvent {
	ch := make(chan PresenceEvent, 16)
	b.mu.Lock()
	b.subscribers[key] = append(b.subscribers[key], ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes ch from key's subscriber list and closes the channel.
func (b *Broker) Unsubscribe(key string, ch chan PresenceEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[key]
	for i, s := range subs {
		if s == ch {
			b.subscribers[key] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(b.subscribers[key]) == 0 {
		delete(b.subscribers, key)
	}
	close(ch)
}

// Publish sends event to every subscriber of key.
// Non-blocking: slow subscribers are skipped if their buffer is full.
func (b *Broker) Publish(key string, event PresenceEvent) {
	b.mu.RLock()
	subs := b.subscribers[key]
	b.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}
