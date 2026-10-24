package chat

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jobba/backend/internal/middleware"
	"github.com/jobba/backend/internal/shared"
)

type streamRecorder struct {
	header http.Header
	status int
	writes chan string
}

func newStreamRecorder() *streamRecorder {
	return &streamRecorder{header: make(http.Header), writes: make(chan string, 8)}
}
func (r *streamRecorder) Header() http.Header    { return r.header }
func (r *streamRecorder) WriteHeader(status int) { r.status = status }
func (r *streamRecorder) Write(body []byte) (int, error) {
	r.writes <- string(body)
	return len(body), nil
}
func (r *streamRecorder) Flush() {}

type streamAuthenticator struct{}

func (streamAuthenticator) Authenticate(context.Context, string) (shared.Actor, error) {
	return shared.Actor{ID: "user-a"}, nil
}

func TestEventStreamRequiresBearerAuthentication(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(nil, NewMemoryHub(HubOptions{})).Register(mux, func(next http.Handler) http.Handler {
		return middleware.Authenticate(streamAuthenticator{}, next)
	})
	recorder := newStreamRecorder()
	mux.ServeHTTP(recorder, mustRequest(t, context.Background()))
	if recorder.status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.status)
	}
}

func mustRequest(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/chat/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestEventStreamFramesPublishedEvents(t *testing.T) {
	hub := NewMemoryHub(HubOptions{})
	handler := NewHandler(nil, hub)
	ctx, cancel := context.WithCancel(shared.WithActor(context.Background(), shared.Actor{ID: "user-a"}))
	defer cancel()
	request := mustRequest(t, ctx)
	recorder := newStreamRecorder()
	done := make(chan struct{})
	go func() { handler.streamEvents(recorder, request); close(done) }()
	select {
	case initial := <-recorder.writes:
		if initial != ": connected\n\n" {
			t.Fatalf("unexpected initial frame: %q", initial)
		}
	case <-time.After(time.Second):
		t.Fatal("stream did not connect")
	}
	if recorder.status != http.StatusOK || recorder.header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("unexpected response: %d %q", recorder.status, recorder.header.Get("Content-Type"))
	}
	hub.Publish(context.Background(), Event{Type: "message:new", UserIDs: []string{"user-a"}, ConversationID: "conversation-a"})
	select {
	case frame := <-recorder.writes:
		if !strings.Contains(frame, "event: message:new\n") || !strings.Contains(frame, `"conversationId":"conversation-a"`) {
			t.Fatalf("unexpected SSE frame: %q", frame)
		}
	case <-time.After(time.Second):
		t.Fatal("stream did not receive event")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stream did not disconnect")
	}
}
