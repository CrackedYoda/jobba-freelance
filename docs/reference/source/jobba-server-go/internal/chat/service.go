package chat

import (
	"context"
	"errors"
	"github.com/jobba/backend/internal/shared"
	"strings"
)

type service struct {
	repo Repository
	hub  Hub
}

func NewService(r Repository, h Hub) Service {
	if h == nil {
		h = NopHub{}
	}
	return &service{r, h}
}
func norm(in ListInput, def int) ListInput {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.Limit < 1 {
		in.Limit = def
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	return in
}
func wrap(e error) error {
	if e == nil {
		return nil
	}
	var a *shared.AppError
	if errors.As(e, &a) {
		return a
	}
	return ErrOperation.WithError(e)
}
func (s *service) Create(c context.Context, u, p, j string) (Conversation, error) {
	if u == p || p == "" || j == "" {
		return Conversation{}, ErrInvalid
	}
	v, e := s.repo.Create(c, u, p, j)
	return v, wrap(e)
}
func (s *service) List(c context.Context, u string, i ListInput) (Page[Conversation], error) {
	v, e := s.repo.List(c, u, norm(i, 20))
	return v, wrap(e)
}
func (s *service) Search(c context.Context, u string, i ListInput) (Page[Conversation], error) {
	i.Query = strings.TrimSpace(i.Query)
	if i.Query == "" || len(i.Query) > 100 {
		return Page[Conversation]{}, ErrInvalid
	}
	v, e := s.repo.List(c, u, norm(i, 20))
	return v, wrap(e)
}
func (s *service) Get(c context.Context, u, id string) (Conversation, error) {
	v, e := s.repo.Get(c, u, id)
	return v, wrap(e)
}
func (s *service) Send(c context.Context, u, id string, i SendInput) (Message, error) {
	i.Content = strings.TrimSpace(i.Content)
	i.FileURL = strings.TrimSpace(i.FileURL)
	if i.MessageType == "" {
		i.MessageType = "text"
	}
	valid := i.MessageType == "text" && i.Content != "" && i.FileURL == "" || (i.MessageType == "image" || i.MessageType == "file" || i.MessageType == "audio") && i.FileURL != ""
	if !valid || len(i.Content) > 10000 {
		return Message{}, ErrInvalid
	}
	v, e := s.repo.Send(c, u, id, i)
	if e == nil {
		users := []string{u}
		if receiver, ok := v.Receiver.(string); ok && receiver != "" && receiver != u {
			users = append(users, receiver)
		}
		s.hub.Publish(c, Event{Type: "message:new", UserIDs: users, ConversationID: id, Data: v})
	}
	return v, wrap(e)
}
func (s *service) Messages(c context.Context, u, id string, i ListInput) (Page[Message], error) {
	v, e := s.repo.Messages(c, u, id, norm(i, 50))
	return v, wrap(e)
}
func (s *service) MarkRead(c context.Context, u, id string) (int64, error) {
	v, e := s.repo.MarkRead(c, u, id)
	if e == nil && v > 0 {
		users := []string{u}
		if conversation, getErr := s.repo.Get(c, u, id); getErr == nil {
			users = users[:0]
			for _, participant := range conversation.Participants {
				if participant.ID != "" {
					users = append(users, participant.ID)
				}
			}
		}
		s.hub.Publish(c, Event{Type: "messages:read", UserIDs: users, ConversationID: id, Data: map[string]any{"readBy": u, "count": v}})
	}
	return v, wrap(e)
}
func (s *service) Delete(c context.Context, u, id string) (Message, error) {
	v, e := s.repo.Delete(c, u, id)
	return v, wrap(e)
}
func (s *service) SetArchived(c context.Context, u, id string, b bool) (Conversation, error) {
	v, e := s.repo.SetArchived(c, u, id, b)
	return v, wrap(e)
}
func (s *service) MarkUrgent(c context.Context, u, id string) (Conversation, error) {
	v, e := s.repo.MarkUrgent(c, u, id)
	return v, wrap(e)
}
func (s *service) Unread(c context.Context, u string) (int64, error) {
	v, e := s.repo.Unread(c, u)
	return v, wrap(e)
}
