package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jobba/backend/internal/shared"
)

type fakeService struct {
	Service
	input ListInput
}

func (f *fakeService) WorkerApplications(_ context.Context, _ string, in ListInput) (Page, error) {
	f.input = in
	return Page{Docs: []Application{}}, nil
}
func TestWorkerListLegacyEnvelopeAndDefaults(t *testing.T) {
	service := &fakeService{}
	h := NewHandler(service)
	mux := http.NewServeMux()
	h.Register(mux)
	r := httptest.NewRequest("GET", "/api/v1/jobs/worker/applications", nil)
	r = r.WithContext(shared.WithActor(r.Context(), shared.Actor{ID: "90b4ab1e-81b3-4e78-9d50-a900239370b2", Role: "worker"}))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"message":"Applications retrieved successfully"`) || !strings.Contains(w.Body.String(), `"docs":[]`) {
		t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
	}
	if service.input.Page != 1 || service.input.Limit != 10 {
		t.Fatalf("unexpected defaults %#v", service.input)
	}
}
func TestApplicationRouteRoleGate(t *testing.T) {
	h := NewHandler(&fakeService{})
	mux := http.NewServeMux()
	h.Register(mux)
	r := httptest.NewRequest("GET", "/api/v1/jobs/worker/applications", nil)
	r = r.WithContext(shared.WithActor(r.Context(), shared.Actor{ID: "90b4ab1e-81b3-4e78-9d50-a900239370b2", Role: "client"}))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 403 || !strings.Contains(w.Body.String(), `"success":false`) {
		t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
	}
}
