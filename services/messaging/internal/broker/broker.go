package broker

import "sync"

// SSEEvent is the payload published to all channel subscribers.
type SSEEvent struct {
	MessageID string `json:"id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

// Broker is an in-memory pub/sub hub for SSE events. Safe for concurrent use.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan SSEEvent // key = channelID
}

func NewBroker() *Broker {
	return &Broker{subscribers: make(map[string][]chan SSEEvent)}
}

// Subscribe registers a new subscriber for channelID.
// The caller MUST call Unsubscribe(channelID, ch) when done (on client disconnect).
func (b *Broker) Subscribe(channelID string) chan SSEEvent {
	ch := make(chan SSEEvent, 16)
	b.mu.Lock()
	b.subscribers[channelID] = append(b.subscribers[channelID], ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes ch from channelID's subscriber list and closes the channel.
func (b *Broker) Unsubscribe(channelID string, ch chan SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[channelID]
	for i, s := range subs {
		if s == ch {
			b.subscribers[channelID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(b.subscribers[channelID]) == 0 {
		delete(b.subscribers, channelID)
	}
	close(ch)
}

// Publish sends event to every subscriber of channelID.
// Non-blocking: slow subscribers are skipped if their buffer is full.
func (b *Broker) Publish(channelID string, event SSEEvent) {
	b.mu.RLock()
	subs := b.subscribers[channelID]
	b.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}
