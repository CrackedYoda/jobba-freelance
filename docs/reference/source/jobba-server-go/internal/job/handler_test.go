package job

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
	filter ListFilter
}

func (f *fakeService) Search(_ context.Context, in ListFilter) (Page, error) {
	f.filter = in
	return Page{Data: []Job{}}, nil
}
func TestSearchRejectsUnknownSort(t *testing.T) {
	h := NewHandler(&fakeService{})
	mux := http.NewServeMux()
	h.Register(mux)
	r := httptest.NewRequest("GET", "/api/v1/jobs/worker/search?sortBy=DROP%20TABLE", nil)
	r = r.WithContext(shared.WithActor(r.Context(), shared.Actor{ID: "90b4ab1e-81b3-4e78-9d50-a900239370b2", Role: "worker"}))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"success":false`) {
		t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
	}
}
func TestSearchRequiresActor(t *testing.T) {
	h := NewHandler(&fakeService{})
	mux := http.NewServeMux()
	h.Register(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/jobs/worker/search", nil))
	if w.Code != 401 || !strings.Contains(w.Body.String(), `"message":"Authentication required"`) {
		t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
	}
}
