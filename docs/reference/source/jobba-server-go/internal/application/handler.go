package application

import (
	"net/http"

	"github.com/jobba/backend/internal/middleware"
	"github.com/jobba/backend/internal/shared"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
func (h *Handler) Register(mux *http.ServeMux) {
	h.RegisterProtected(mux, func(next http.Handler) http.Handler { return next })
}
func (h *Handler) RegisterProtected(mux *http.ServeMux, protect func(http.Handler) http.Handler) {
	register := func(pattern, role string, next http.HandlerFunc) { mux.Handle(pattern, protect(guard(role, next))) }
	register("POST /api/v1/jobs/worker/{jobId}/apply", "worker", h.apply)
	register("GET /api/v1/jobs/worker/applications", "worker", h.workerList)
	register("PUT /api/v1/jobs/client/application/{applicationId}/accept", "client", h.accept)
}

func (h *Handler) HandleJobApplications(w http.ResponseWriter, r *http.Request) { h.jobList(w, r) }
func guard(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, ok := shared.ActorFromContext(r.Context())
		if !ok {
			fail(w, r, ErrUnauthorized)
			return
		}
		if a.Role != role {
			fail(w, r, ErrForbidden)
			return
		}
		next(w, r)
	}
}
func fail(w http.ResponseWriter, r *http.Request, err error) {
	shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), err)
}
func bad(w http.ResponseWriter, message string) { shared.WriteLegacyJSON(w, 400, message, nil) }
func actor(r *http.Request) shared.Actor        { a, _ := shared.ActorFromContext(r.Context()); return a }
func (h *Handler) apply(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("jobId")
	if !validID(id) {
		bad(w, "invalid job ID")
		return
	}
	var in CreateInput
	if err := decode(r, &in); err != nil {
		bad(w, "invalid request body")
		return
	}
	a, err := h.service.Apply(r.Context(), actor(r).ID, id, in)
	if err != nil {
		fail(w, r, err)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Application submitted successfully", a)
}
func (h *Handler) workerList(w http.ResponseWriter, r *http.Request) {
	in, err := listInput(r)
	if err != nil {
		bad(w, err.Error())
		return
	}
	p, err := h.service.WorkerApplications(r.Context(), actor(r).ID, in)
	if err != nil {
		fail(w, r, err)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Applications retrieved successfully", p)
}
func (h *Handler) jobList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("jobId")
	if !validID(id) {
		bad(w, "invalid job ID")
		return
	}
	in, err := listInput(r)
	if err != nil {
		bad(w, err.Error())
		return
	}
	p, err := h.service.JobApplications(r.Context(), actor(r).ID, id, in)
	if err != nil {
		fail(w, r, err)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Job applications retrieved successfully", p)
}
func (h *Handler) accept(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("applicationId")
	if !validID(id) {
		bad(w, "invalid application ID")
		return
	}
	a, err := h.service.Accept(r.Context(), actor(r).ID, id)
	if err != nil {
		fail(w, r, err)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Application accepted and job assigned successfully", a)
}
