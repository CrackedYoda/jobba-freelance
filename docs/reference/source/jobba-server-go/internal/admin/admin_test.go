package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jobba/backend/internal/shared"
)

func TestAdminRoutesRejectNonAdminBeforeRepository(t *testing.T) {
	handler := NewHandler(&Repository{})
	mux := http.NewServeMux()
	handler.RegisterProtected(mux, func(next http.Handler) http.Handler { return next })
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	request = request.WithContext(shared.WithActor(request.Context(), shared.Actor{ID: "worker-1", Role: "worker"}))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("got %d", response.Code)
	}
}
func TestRolePermissionsRequireResourcesObject(t *testing.T) {
	if validRole(roleBody{Name: "ops", Permissions: []byte(`{"anything":[]}`)}) {
		t.Fatal("permissions without resources accepted")
	}
	if !validRole(roleBody{Name: "ops", Permissions: []byte(`{"resources":[]}`)}) {
		t.Fatal("valid permissions rejected")
	}
}
