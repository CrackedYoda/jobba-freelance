package job

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
	mux.HandleFunc("GET /api/v1/jobs", h.publicList)
	register := func(pattern, role string, next http.HandlerFunc) { mux.Handle(pattern, protect(roleGuard(role, next))) }
	register("POST /api/v1/jobs/client", "client", h.create)
	register("GET /api/v1/jobs/client", "client", h.listClient)
	register("GET /api/v1/jobs/client/statistics", "client", h.statistics)
	register("GET /api/v1/jobs/client/{jobId}", "client", h.getClient)
	register("PUT /api/v1/jobs/client/{jobId}", "client", h.update)
	register("DELETE /api/v1/jobs/client/{jobId}", "client", h.delete)
	register("PUT /api/v1/jobs/client/{jobId}/close", "client", h.close)
	register("PUT /api/v1/jobs/client/{jobId}/complete", "client", h.complete)
	register("PUT /api/v1/jobs/client/{jobId}/assign", "client", h.assign)
	register("GET /api/v1/jobs/worker", "worker", h.listWorker)
	register("GET /api/v1/jobs/worker/search", "worker", h.search)
	register("GET /api/v1/jobs/worker/saved", "worker", h.saved)
	register("GET /api/v1/jobs/worker/category/{category}", "worker", h.category)
	register("GET /api/v1/jobs/worker/{jobId}", "worker", h.getWorker)
	register("POST /api/v1/jobs/worker/{jobId}/save", "worker", h.save)
	register("DELETE /api/v1/jobs/worker/{jobId}/save", "worker", h.unsave)
}

func roleGuard(role string, next http.HandlerFunc) http.HandlerFunc {
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
func actor(r *http.Request) (shared.Actor, error) {
	a, ok := shared.ActorFromContext(r.Context())
	if !ok {
		return a, ErrUnauthorized
	}
	return a, nil
}
func fail(w http.ResponseWriter, r *http.Request, e error) {
	shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), e)
}
func bad(w http.ResponseWriter, message string) {
	shared.WriteLegacyJSON(w, http.StatusBadRequest, message, nil)
}
func pathID(r *http.Request) (string, bool) { id := r.PathValue("jobId"); return id, validID(id) }
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	var in CreateInput
	if e = decodeJSON(r, &in); e != nil {
		bad(w, e.Error())
		return
	}
	j, e := h.service.Create(r.Context(), a.ID, in)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, http.StatusCreated, "Job created successfully", j)
}
func (h *Handler) listClient(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	f, e := listFilterFromRequest(r)
	if e != nil {
		bad(w, e.Error())
		return
	}
	p, e := h.service.ListClient(r.Context(), a.ID, f)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Jobs retrieved successfully", p)
}
func (h *Handler) listWorker(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	f, e := listFilterFromRequest(r)
	if e != nil {
		bad(w, e.Error())
		return
	}
	p, e := h.service.ListWorker(r.Context(), a.ID, f)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Jobs retrieved successfully", p.Legacy(f.Page, f.Limit))
}
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	if _, e := actor(r); e != nil {
		fail(w, r, e)
		return
	}
	f, e := listFilterFromRequest(r)
	if e != nil {
		bad(w, e.Error())
		return
	}
	p, e := h.service.Search(r.Context(), f)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Jobs retrieved successfully", p.Legacy(f.Page, f.Limit))
}

func (h *Handler) publicList(w http.ResponseWriter, r *http.Request) {
	f, e := listFilterFromRequest(r)
	if e != nil {
		bad(w, e.Error())
		return
	}
	p, e := h.service.Search(r.Context(), f)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, http.StatusOK, "Jobs retrieved successfully", p)
}
func (h *Handler) category(w http.ResponseWriter, r *http.Request) {
	if _, e := actor(r); e != nil {
		fail(w, r, e)
		return
	}
	p, e := paginationFromRequest(r)
	if e != nil {
		bad(w, e.Error())
		return
	}
	v, e := h.service.ByCategory(r.Context(), r.PathValue("category"), p)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Jobs retrieved successfully", v.Legacy(p.Page, p.Limit))
}
func (h *Handler) getClient(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	j, e := h.service.GetClient(r.Context(), id, a.ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Job retrieved successfully", j)
}
func (h *Handler) getWorker(w http.ResponseWriter, r *http.Request) {
	if _, e := actor(r); e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	a, _ := actor(r)
	j, e := h.service.GetWorker(r.Context(), id, a.ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Job details retrieved successfully", j)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	var in UpdateInput
	if e = decodeJSON(r, &in); e != nil {
		bad(w, e.Error())
		return
	}
	if in.empty() {
		bad(w, "request body must include at least one job field")
		return
	}
	j, e := h.service.Update(r.Context(), id, a.ID, in)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Job updated successfully", j)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	if e = h.service.Delete(r.Context(), id, a.ID); e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Job deleted successfully", nil)
}
func (h *Handler) close(w http.ResponseWriter, r *http.Request) {
	h.state(w, r, "Job closed successfully", func(c string, id string) (Job, error) { return h.service.Close(r.Context(), id, c) })
}
func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	h.state(w, r, "Job completed successfully", func(c string, id string) (Job, error) { return h.service.Complete(r.Context(), id, c) })
}
func (h *Handler) state(w http.ResponseWriter, r *http.Request, message string, fn func(string, string) (Job, error)) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	j, e := fn(a.ID, id)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, message, j)
}
func (h *Handler) assign(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	var in assignInput
	if e = decodeJSON(r, &in); e != nil || !validID(in.WorkerID) {
		bad(w, "Worker ID is required")
		return
	}
	j, e := h.service.Assign(r.Context(), id, a.ID, in.WorkerID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Worker assigned successfully", j)
}
func (h *Handler) statistics(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	v, e := h.service.Statistics(r.Context(), a.ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Statistics retrieved successfully", v)
}
func (h *Handler) save(w http.ResponseWriter, r *http.Request)   { h.savedMutation(w, r, true) }
func (h *Handler) unsave(w http.ResponseWriter, r *http.Request) { h.savedMutation(w, r, false) }
func (h *Handler) savedMutation(w http.ResponseWriter, r *http.Request, save bool) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	id, ok := pathID(r)
	if !ok {
		bad(w, "invalid job ID")
		return
	}
	message := "Job removed from saved list"
	if save {
		e = h.service.Save(r.Context(), a.ID, id)
		message = "Job saved successfully"
	} else {
		e = h.service.Unsave(r.Context(), a.ID, id)
	}
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, message, nil)
}
func (h *Handler) saved(w http.ResponseWriter, r *http.Request) {
	a, e := actor(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	p, e := paginationFromRequest(r)
	if e != nil {
		bad(w, e.Error())
		return
	}
	v, e := h.service.Saved(r.Context(), a.ID, p)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Saved jobs retrieved successfully", v.Data)
}
