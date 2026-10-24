package chat

import (
	"context"
	"errors"
)

type Repository interface {
	Create(context.Context, string, string, string) (Conversation, error)
	List(context.Context, string, ListInput) (Page[Conversation], error)
	Get(context.Context, string, string) (Conversation, error)
	Send(context.Context, string, string, SendInput) (Message, error)
	Messages(context.Context, string, string, ListInput) (Page[Message], error)
	MarkRead(context.Context, string, string) (int64, error)
	Delete(context.Context, string, string) (Message, error)
	SetArchived(context.Context, string, string, bool) (Conversation, error)
	MarkUrgent(context.Context, string, string) (Conversation, error)
	Unread(context.Context, string) (int64, error)
}
type Service interface {
	Create(context.Context, string, string, string) (Conversation, error)
	List(context.Context, string, ListInput) (Page[Conversation], error)
	Search(context.Context, string, ListInput) (Page[Conversation], error)
	Get(context.Context, string, string) (Conversation, error)
	Send(context.Context, string, string, SendInput) (Message, error)
	Messages(context.Context, string, string, ListInput) (Page[Message], error)
	MarkRead(context.Context, string, string) (int64, error)
	Delete(context.Context, string, string) (Message, error)
	SetArchived(context.Context, string, string, bool) (Conversation, error)
	MarkUrgent(context.Context, string, string) (Conversation, error)
	Unread(context.Context, string) (int64, error)
}
type Hub interface{ Publish(context.Context, Event) }
type EventSubscriber interface {
	Subscribe(string) (<-chan Event, func(), error)
}
type NopHub struct{}

func (NopHub) Publish(context.Context, Event) {}
func (NopHub) Subscribe(string) (<-chan Event, func(), error) {
	return nil, func() {}, errors.New("chat event stream unavailable")
}
