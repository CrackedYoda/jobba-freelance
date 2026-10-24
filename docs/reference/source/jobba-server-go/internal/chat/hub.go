package chat

import (
	"context"
	"errors"
	"sync"
)

var ErrSubscriberLimit = errors.New("chat event subscriber limit reached")

type HubOptions struct {
	QueueSize          int
	MaxSubscribers     int
	MaxSubscribersUser int
}

type MemoryHub struct {
	mu                                            sync.RWMutex
	byUser                                        map[string]map[uint64]chan Event
	nextID                                        uint64
	total                                         int
	queueSize, maxSubscribers, maxSubscribersUser int
}

func NewMemoryHub(options HubOptions) *MemoryHub {
	if options.QueueSize <= 0 {
		options.QueueSize = 32
	}
	if options.MaxSubscribers <= 0 {
		options.MaxSubscribers = 1000
	}
	if options.MaxSubscribersUser <= 0 {
		options.MaxSubscribersUser = 5
	}
	return &MemoryHub{byUser: make(map[string]map[uint64]chan Event), queueSize: options.QueueSize, maxSubscribers: options.MaxSubscribers, maxSubscribersUser: options.MaxSubscribersUser}
}

func (h *MemoryHub) Subscribe(userID string) (<-chan Event, func(), error) {
	if userID == "" {
		return nil, func() {}, errors.New("user id is required")
	}
	h.mu.Lock()
	if h.total >= h.maxSubscribers || len(h.byUser[userID]) >= h.maxSubscribersUser {
		h.mu.Unlock()
		return nil, func() {}, ErrSubscriberLimit
	}
	h.nextID++
	id := h.nextID
	ch := make(chan Event, h.queueSize)
	if h.byUser[userID] == nil {
		h.byUser[userID] = make(map[uint64]chan Event)
	}
	h.byUser[userID][id] = ch
	h.total++
	h.mu.Unlock()
	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			if subscriptions := h.byUser[userID]; subscriptions != nil {
				if existing, ok := subscriptions[id]; ok {
					delete(subscriptions, id)
					h.total--
					close(existing)
				}
				if len(subscriptions) == 0 {
					delete(h.byUser, userID)
				}
			}
			h.mu.Unlock()
		})
	}
	return ch, cancel, nil
}

func (h *MemoryHub) Publish(_ context.Context, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	seen := make(map[string]struct{}, len(event.UserIDs))
	for _, userID := range event.UserIDs {
		if _, duplicate := seen[userID]; duplicate {
			continue
		}
		seen[userID] = struct{}{}
		for _, ch := range h.byUser[userID] {
			select {
			case ch <- event:
			default:
			}
		}
	}
}
