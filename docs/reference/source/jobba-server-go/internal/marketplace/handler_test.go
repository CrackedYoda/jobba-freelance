package marketplace

import (
	"context"
	"github.com/jobba/backend/internal/shared"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeService struct {
	Service
	offerUser string
	offerRole string
}

func (f *fakeService) Recommended(context.Context, string, int) ([]Job, error) { return []Job{}, nil }
func (f *fakeService) Offers(_ context.Context, user, role string, _ OfferQuery) (OfferPage, error) {
	f.offerUser, f.offerRole = user, role
	return OfferPage{Docs: []Offer{}}, nil
}
func TestRecommendedLegacyEnvelope(t *testing.T) {
	h := NewHandler(&fakeService{})
	m := http.NewServeMux()
	h.RegisterProtected(m, func(n http.Handler) http.Handler { return n })
	r := httptest.NewRequest("GET", "/api/v1/jobs/worker/recommended", nil)
	r = r.WithContext(shared.WithActor(r.Context(), shared.Actor{ID: "worker", Role: "worker"}))
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"message":"Recommended jobs retrieved successfully"`) {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
func TestRecommendedRejectsClient(t *testing.T) {
	h := NewHandler(&fakeService{})
	m := http.NewServeMux()
	h.RegisterProtected(m, func(n http.Handler) http.Handler { return n })
	r := httptest.NewRequest("GET", "/api/v1/jobs/worker/recommended", nil)
	r = r.WithContext(shared.WithActor(r.Context(), shared.Actor{ID: "client", Role: "client"}))
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestWorkerOfferListUsesAuthenticatedScope(t *testing.T) {
	service := &fakeService{}
	h := NewHandler(service)
	m := http.NewServeMux()
	h.RegisterProtected(m, func(n http.Handler) http.Handler { return n })
	r := httptest.NewRequest("GET", "/api/v1/jobs/worker/offers", nil)
	r = r.WithContext(shared.WithActor(r.Context(), shared.Actor{ID: "worker-1", Role: "worker"}))
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
	if w.Code != 200 || service.offerUser != "worker-1" || service.offerRole != "worker" {
		t.Fatalf("scope status=%d user=%q role=%q", w.Code, service.offerUser, service.offerRole)
	}
}
