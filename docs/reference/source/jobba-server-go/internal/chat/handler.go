package chat

import (
	"encoding/json"
	"github.com/jobba/backend/internal/middleware"
	"github.com/jobba/backend/internal/shared"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	s         Service
	events    EventSubscriber
	keepAlive time.Duration
}

func NewHandler(s Service, events ...EventSubscriber) *Handler {
	h := &Handler{s: s, keepAlive: 20 * time.Second}
	if len(events) > 0 {
		h.events = events[0]
	}
	return h
}
func (h *Handler) Register(m *http.ServeMux, p func(http.Handler) http.Handler) {
	reg := func(x string, f http.HandlerFunc) { m.Handle(x, p(http.HandlerFunc(f))) }
	reg("POST /api/v1/chat/conversations", h.create)
	reg("GET /api/v1/chat/conversations", h.list)
	reg("GET /api/v1/chat/conversations/search", h.search)
	reg("GET /api/v1/chat/conversations/{conversationId}", h.get)
	reg("GET /api/v1/chat/conversations/{conversationId}/messages", h.messages)
	reg("POST /api/v1/chat/conversations/{conversationId}/messages", h.send)
	reg("POST /api/v1/chat/conversations/{conversationId}/read", h.read)
	reg("POST /api/v1/chat/conversations/{conversationId}/archive", h.archive)
	reg("POST /api/v1/chat/conversations/{conversationId}/unarchive", h.unarchive)
	reg("POST /api/v1/chat/conversations/{conversationId}/urgent", h.urgent)
	reg("DELETE /api/v1/chat/messages/{messageId}", h.del)
	reg("GET /api/v1/chat/unread-count", h.unread)
	reg("GET /api/v1/chat/events", h.streamEvents)
}

func (h *Handler) streamEvents(w http.ResponseWriter, r *http.Request) {
	if h.events == nil {
		http.Error(w, "chat event stream unavailable", http.StatusServiceUnavailable)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	events, unsubscribe, err := h.events.Subscribe(user(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer unsubscribe()
	// The server's ordinary response write deadline is intentionally finite. An
	// authenticated event stream owns its lifecycle through request cancellation.
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(": connected\n\n"))
	flusher.Flush()
	ticker := time.NewTicker(h.keepAlive)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case event, open := <-events:
			if !open {
				return
			}
			payload, marshalErr := json.Marshal(event)
			if marshalErr != nil {
				continue
			}
			if _, err = w.Write([]byte("event: " + event.Type + "\ndata: " + string(payload) + "\n\n")); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err = w.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
func user(r *http.Request) string { a, _ := shared.ActorFromContext(r.Context()); return a.ID }
func fail(w http.ResponseWriter, r *http.Request, e error) {
	shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), e)
}
func bad(w http.ResponseWriter, s string) { shared.WriteLegacyJSON(w, 400, s, nil) }
func listIn(r *http.Request) (ListInput, error) {
	q := r.URL.Query()
	p, _ := strconv.Atoi(q.Get("page"))
	l, _ := strconv.Atoi(q.Get("limit"))
	i := ListInput{Page: p, Limit: l, Query: q.Get("q")}
	for n, d := range map[string]**bool{"isArchived": &i.Archived, "isActive": &i.Active} {
		if v := q.Get(n); v != "" {
			b, e := strconv.ParseBool(v)
			if e != nil {
				return i, e
			}
			*d = &b
		}
	}
	return i, nil
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Participant string `json:"participantId"`
		Job         string `json:"jobId"`
	}
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		bad(w, "invalid request body")
		return
	}
	c, e := h.s.Create(r.Context(), user(r), v.Participant, v.Job)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Conversation created successfully", c)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	i, e := listIn(r)
	if e != nil {
		bad(w, "invalid query parameters")
		return
	}
	v, e := h.s.List(r.Context(), user(r), i)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Conversations fetched successfully", v)
}
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	i, e := listIn(r)
	if e != nil || strings.TrimSpace(i.Query) == "" {
		bad(w, "Search query is required")
		return
	}
	v, e := h.s.Search(r.Context(), user(r), i)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Search completed successfully", v.Docs)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.Get(r.Context(), user(r), r.PathValue("conversationId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Conversation fetched successfully", v)
}
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	var i SendInput
	if json.NewDecoder(r.Body).Decode(&i) != nil {
		bad(w, "invalid request body")
		return
	}
	v, e := h.s.Send(r.Context(), user(r), r.PathValue("conversationId"), i)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Message sent successfully", v)
}
func (h *Handler) messages(w http.ResponseWriter, r *http.Request) {
	i, e := listIn(r)
	if e != nil {
		bad(w, "invalid query parameters")
		return
	}
	v, e := h.s.Messages(r.Context(), user(r), r.PathValue("conversationId"), i)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Messages fetched successfully", v)
}
func (h *Handler) read(w http.ResponseWriter, r *http.Request) {
	_, e := h.s.MarkRead(r.Context(), user(r), r.PathValue("conversationId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Messages marked as read successfully", nil)
}
func (h *Handler) del(w http.ResponseWriter, r *http.Request) {
	_, e := h.s.Delete(r.Context(), user(r), r.PathValue("messageId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Message deleted successfully", nil)
}
func (h *Handler) setArchive(w http.ResponseWriter, r *http.Request, b bool) {
	v, e := h.s.SetArchived(r.Context(), user(r), r.PathValue("conversationId"), b)
	if e != nil {
		fail(w, r, e)
		return
	}
	msg := "Conversation archived successfully"
	if !b {
		msg = "Conversation unarchived successfully"
	}
	shared.WriteLegacyJSON(w, 200, msg, v)
}
func (h *Handler) archive(w http.ResponseWriter, r *http.Request)   { h.setArchive(w, r, true) }
func (h *Handler) unarchive(w http.ResponseWriter, r *http.Request) { h.setArchive(w, r, false) }
func (h *Handler) urgent(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.MarkUrgent(r.Context(), user(r), r.PathValue("conversationId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Conversation marked as urgent successfully", v)
}
func (h *Handler) unread(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.Unread(r.Context(), user(r))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Unread count fetched successfully", map[string]int64{"unreadCount": v})
}
