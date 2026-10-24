package marketplace

import (
	"encoding/json"
	"github.com/jobba/backend/internal/middleware"
	"github.com/jobba/backend/internal/shared"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ s Service }

func NewHandler(s Service) *Handler { return &Handler{s} }
func (h *Handler) RegisterProtected(m *http.ServeMux, protect func(http.Handler) http.Handler) {
	reg := func(p, role string, f http.HandlerFunc) { m.Handle(p, protect(roleGuard(role, f))) }
	regAny := func(p string, f http.HandlerFunc) { m.Handle(p, protect(authGuard(f))) }
	reg("GET /api/v1/jobs/client/workers", "client", h.searchWorkers)
	reg("GET /api/v1/jobs/client/workers/saved-workers", "client", h.savedWorkers)
	reg("POST /api/v1/jobs/client/workers/{workerId}/save", "client", h.toggleSaved)
	reg("GET /api/v1/jobs/worker/recommended", "worker", h.recommended)
	reg("POST /api/v1/jobs/client/offers", "client", h.createOffer)
	reg("GET /api/v1/jobs/client/offers", "client", h.offers)
	reg("GET /api/v1/jobs/client/offers/{offerId}", "client", h.offer)
	reg("GET /api/v1/jobs/worker/offers", "worker", h.offers)
	reg("GET /api/v1/jobs/worker/offers/{offerId}", "worker", h.offer)
	reg("POST /api/v1/jobs/worker/offers/{offerId}/accept", "worker", h.acceptOffer)
	reg("POST /api/v1/jobs/worker/offers/{offerId}/reject", "worker", h.rejectOffer)
	reg("POST /api/v1/jobs/client/{jobId}/review", "client", h.createReview)
	reg("POST /api/v1/jobs/worker/{jobId}/review", "worker", h.createReview)
	reg("GET /api/v1/jobs/client/reviews/{userId}", "client", h.reviews)
	reg("GET /api/v1/jobs/worker/reviews/{userId}", "worker", h.reviews)
	regAny("GET /api/v1/jobs/context/{jobId}", h.jobContext)
}
func (h *Handler) HandleWorkerProfile(w http.ResponseWriter, r *http.Request) { h.worker(w, r) }
func authGuard(n http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := shared.ActorFromContext(r.Context()); !ok {
			fail(w, r, ErrUnauthorized)
			return
		}
		n(w, r)
	}
}
func roleGuard(role string, n http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, ok := shared.ActorFromContext(r.Context())
		if !ok {
			fail(w, r, ErrUnauthorized)
			return
		}
		if a.Role != role && !(strings.HasPrefix(r.URL.Path, "/api/v1/workers/") && (a.Role == "admin" || a.Role == "worker" || a.Role == "client")) {
			fail(w, r, ErrForbidden)
			return
		}
		n(w, r)
	}
}
func fail(w http.ResponseWriter, r *http.Request, e error) {
	shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), e)
}
func actor(r *http.Request) shared.Actor { a, _ := shared.ActorFromContext(r.Context()); return a }
func decode(r *http.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func search(r *http.Request, saved string) (Search, error) {
	q := r.URL.Query()
	p := Search{Page: 1, Limit: 10, Search: strings.TrimSpace(q.Get("search")), SkillID: q.Get("skillId"), WorkType: q.Get("workType"), Country: q.Get("country"), State: q.Get("stateOfResidence"), LGA: q.Get("lga"), SavedBy: saved}
	var e error
	if q.Get("page") != "" {
		p.Page, e = strconv.Atoi(q.Get("page"))
	}
	if e == nil && q.Get("limit") != "" {
		p.Limit, e = strconv.Atoi(q.Get("limit"))
	}
	if e != nil || p.Page < 1 || p.Limit < 1 || p.Limit > 100 {
		return p, ErrInvalid
	}
	return p, nil
}
func (h *Handler) searchWorkers(w http.ResponseWriter, r *http.Request) {
	q, e := search(r, "")
	if e != nil {
		fail(w, r, e)
		return
	}
	v, e := h.s.SearchWorkers(r.Context(), q)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Workers retrieved successfully", v)
}
func (h *Handler) savedWorkers(w http.ResponseWriter, r *http.Request) {
	q, e := search(r, actor(r).ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	v, e := h.s.SearchWorkers(r.Context(), q)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Saved workers retrieved successfully", v)
}
func (h *Handler) worker(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.Worker(r.Context(), r.PathValue("workerId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Worker profile retrieved successfully", v)
}
func (h *Handler) toggleSaved(w http.ResponseWriter, r *http.Request) {
	saved, e := h.s.ToggleSaved(r.Context(), actor(r).ID, r.PathValue("workerId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	msg := "Worker removed from saved list."
	if saved {
		msg = "Worker saved successfully."
	}
	shared.WriteLegacyJSON(w, 200, msg, nil)
}
func profileJSON(w http.ResponseWriter, v PublicProfile) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": v})
}
func (h *Handler) publicProfile(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.Profile(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, r, e)
		return
	}
	profileJSON(w, v)
}
func (h *Handler) myProfile(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.Profile(r.Context(), actor(r).ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	profileJSON(w, v)
}
func (h *Handler) upsertProfile(w http.ResponseWriter, r *http.Request) {
	var in ProfileInput
	if e := decode(r, &in); e != nil {
		fail(w, r, ErrInvalid)
		return
	}
	v, e := h.s.UpsertProfile(r.Context(), actor(r).ID, in)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Profile updated successfully", v)
}
func (h *Handler) recommended(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if x := r.URL.Query().Get("limit"); x != "" {
		var e error
		limit, e = strconv.Atoi(x)
		if e != nil || limit < 1 || limit > 100 {
			fail(w, r, ErrInvalid)
			return
		}
	}
	v, e := h.s.Recommended(r.Context(), actor(r).ID, limit)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Recommended jobs retrieved successfully", v)
}
func (h *Handler) createOffer(w http.ResponseWriter, r *http.Request) {
	var in OfferInput
	if e := decode(r, &in); e != nil {
		fail(w, r, ErrInvalid)
		return
	}
	v, e := h.s.CreateOffer(r.Context(), actor(r).ID, in)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Offer sent safely to the chat", map[string]any{"offer": v, "message": nil})
}
func offerQuery(r *http.Request) (OfferQuery, error) {
	q := OfferQuery{Page: 1, Limit: 10, Status: r.URL.Query().Get("status")}
	var e error
	if v := r.URL.Query().Get("page"); v != "" {
		q.Page, e = strconv.Atoi(v)
	}
	if e == nil {
		if v := r.URL.Query().Get("limit"); v != "" {
			q.Limit, e = strconv.Atoi(v)
		}
	}
	if e != nil {
		return q, ErrInvalid
	}
	return q, nil
}
func (h *Handler) offers(w http.ResponseWriter, r *http.Request) {
	q, e := offerQuery(r)
	if e != nil {
		fail(w, r, e)
		return
	}
	a := actor(r)
	v, e := h.s.Offers(r.Context(), a.ID, a.Role, q)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Offers retrieved successfully", v)
}
func (h *Handler) offer(w http.ResponseWriter, r *http.Request) {
	a := actor(r)
	v, e := h.s.Offer(r.Context(), a.ID, a.Role, r.PathValue("offerId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Offer retrieved successfully", v)
}
func (h *Handler) acceptOffer(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.AcceptOffer(r.Context(), r.PathValue("offerId"), actor(r).ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Offer accepted successfully", v)
}
func (h *Handler) rejectOffer(w http.ResponseWriter, r *http.Request) {
	e := h.s.RejectOffer(r.Context(), r.PathValue("offerId"), actor(r).ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Offer rejected successfully", nil)
}
func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if e := decode(r, &in); e != nil {
		fail(w, r, ErrInvalid)
		return
	}
	v, e := h.s.CreateReview(r.Context(), actor(r).ID, r.PathValue("jobId"), in.Rating, in.Comment)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Review created successfully", v)
}
func (h *Handler) reviews(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.Reviews(r.Context(), r.PathValue("userId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Reviews fetched successfully", v)
}
func (h *Handler) jobContext(w http.ResponseWriter, r *http.Request) {
	a := actor(r)
	v, e := h.s.JobContext(r.Context(), a.ID, a.Role, r.PathValue("jobId"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Job action context retrieved successfully", v)
}
