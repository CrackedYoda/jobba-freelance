package chat

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryHubFanoutCleanupAndLimits(t *testing.T) {
	hub := NewMemoryHub(HubOptions{QueueSize: 1, MaxSubscribers: 2, MaxSubscribersUser: 1})
	a, cancelA, err := hub.Subscribe("a")
	if err != nil {
		t.Fatal(err)
	}
	defer cancelA()
	b, cancelB, err := hub.Subscribe("b")
	if err != nil {
		t.Fatal(err)
	}
	defer cancelB()
	if _, _, err = hub.Subscribe("a"); !errors.Is(err, ErrSubscriberLimit) {
		t.Fatalf("expected per-user limit, got %v", err)
	}
	event := Event{Type: "message:new", UserIDs: []string{"a", "b", "a"}, ConversationID: "c"}
	hub.Publish(context.Background(), event)
	for name, ch := range map[string]<-chan Event{"a": a, "b": b} {
		select {
		case got := <-ch:
			if got.ConversationID != "c" {
				t.Fatalf("%s got %#v", name, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s did not receive event", name)
		}
	}
	cancelA()
	if _, _, err = hub.Subscribe("replacement"); err != nil {
		t.Fatalf("cleanup did not release capacity: %v", err)
	}
}

func TestMemoryHubPublishNeverBlocksWhenQueueFull(t *testing.T) {
	hub := NewMemoryHub(HubOptions{QueueSize: 1, MaxSubscribers: 1, MaxSubscribersUser: 1})
	_, cancel, err := hub.Subscribe("a")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	hub.Publish(context.Background(), Event{UserIDs: []string{"a"}})
	done := make(chan struct{})
	go func() { hub.Publish(context.Background(), Event{UserIDs: []string{"a"}}); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("publish blocked on a full subscriber queue")
	}
}
