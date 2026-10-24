package chat

import (
	"context"
	"testing"
)

type fakeRepo struct {
	Repository
	sent SendInput
}

type captureHub struct{ events []Event }

func (h *captureHub) Publish(_ context.Context, event Event) { h.events = append(h.events, event) }

func (f *fakeRepo) Send(_ context.Context, _, _ string, i SendInput) (Message, error) {
	f.sent = i
	return Message{MessageType: i.MessageType}, nil
}

func TestSendPublishesToBothParticipants(t *testing.T) {
	f := &fakeRepo{}
	h := &captureHub{}
	f.Repository = nil
	s := NewService(&sendTargetRepo{fakeRepo: f}, h)
	if _, err := s.Send(context.Background(), "sender", "conversation", SendInput{Content: "hello"}); err != nil {
		t.Fatal(err)
	}
	if len(h.events) != 1 || len(h.events[0].UserIDs) != 2 || h.events[0].UserIDs[0] != "sender" || h.events[0].UserIDs[1] != "receiver" {
		t.Fatalf("unexpected event: %#v", h.events)
	}
}

type sendTargetRepo struct{ *fakeRepo }

func (r *sendTargetRepo) Send(ctx context.Context, userID, conversationID string, input SendInput) (Message, error) {
	m, err := r.fakeRepo.Send(ctx, userID, conversationID, input)
	m.Receiver = "receiver"
	return m, err
}
func TestSendValidatesPayloadAndDefaultsText(t *testing.T) {
	f := &fakeRepo{}
	s := NewService(f, nil)
	if _, e := s.Send(context.Background(), "u", "c", SendInput{Content: " hi "}); e != nil {
		t.Fatal(e)
	}
	if f.sent.MessageType != "text" || f.sent.Content != "hi" {
		t.Fatalf("unexpected: %#v", f.sent)
	}
	if _, e := s.Send(context.Background(), "u", "c", SendInput{MessageType: "image"}); e == nil {
		t.Fatal("expected invalid attachment")
	}
}
func TestPaginationIsBounded(t *testing.T) {
	i := norm(ListInput{Page: -1, Limit: 1000}, 20)
	if i.Page != 1 || i.Limit != 100 {
		t.Fatalf("unexpected %#v", i)
	}
}
