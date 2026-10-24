package dispute

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobba/backend/internal/middleware"
	"github.com/jobba/backend/internal/shared"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Dispute struct {
	ID                string     `json:"_id"`
	JobID             string     `json:"job"`
	InitiatorID       string     `json:"initiator"`
	ResponderID       string     `json:"responder"`
	Reason            string     `json:"reason"`
	Description       string     `json:"description"`
	EvidenceURLs      []string   `json:"evidenceUrls"`
	Status            string     `json:"status"`
	AdminAssignedID   *string    `json:"adminAssigned,omitempty"`
	ResolutionMessage *string    `json:"resolutionMessage,omitempty"`
	ResolvedAt        *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
type Page struct {
	Docs          []Dispute `json:"docs"`
	TotalDocs     int64     `json:"totalDocs"`
	Limit         int       `json:"limit"`
	TotalPages    int       `json:"totalPages"`
	Page          int       `json:"page"`
	PagingCounter int       `json:"pagingCounter"`
	HasPrevPage   bool      `json:"hasPrevPage"`
	HasNextPage   bool      `json:"hasNextPage"`
	PrevPage      *int      `json:"prevPage"`
	NextPage      *int      `json:"nextPage"`
}

var notFound = &shared.AppError{Code: "DISPUTE_NOT_FOUND", Status: 404, Message: "Dispute not found"}
var forbidden = &shared.AppError{Code: "FORBIDDEN", Status: 403, Message: "Unauthorized to access this dispute"}
var conflict = &shared.AppError{Code: "ACTIVE_DISPUTE_EXISTS", Status: 409, Message: "An active dispute already exists for this job"}
var closed = &shared.AppError{Code: "DISPUTE_CLOSED", Status: 400, Message: "Cannot modify a closed dispute"}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(p *pgxpool.Pool) *Repository { return &Repository{p} }

const cols = `id::text,job_id::text,initiator_id::text,responder_id::text,reason::text,description,evidence_urls,status::text,admin_assigned_id::text,resolution_message,resolved_at,created_at,updated_at`

type scanner interface{ Scan(...any) error }

func scan(s scanner) (Dispute, error) {
	var d Dispute
	e := s.Scan(&d.ID, &d.JobID, &d.InitiatorID, &d.ResponderID, &d.Reason, &d.Description, &d.EvidenceURLs, &d.Status, &d.AdminAssignedID, &d.ResolutionMessage, &d.ResolvedAt, &d.CreatedAt, &d.UpdatedAt)
	return d, e
}
func (r *Repository) Create(ctx context.Context, actor, job, reason, desc string, ev []string) (Dispute, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Dispute{}, e
	}
	defer tx.Rollback(ctx)
	var client string
	var worker *string
	e = tx.QueryRow(ctx, `SELECT client_id::text,assigned_worker_id::text FROM jobs WHERE id=$1::uuid AND is_deleted=false`, job).Scan(&client, &worker)
	if errors.Is(e, pgx.ErrNoRows) {
		return Dispute{}, notFound
	}
	if e != nil {
		return Dispute{}, e
	}
	var responder string
	if actor == client && worker != nil {
		responder = *worker
	} else if worker != nil && actor == *worker {
		responder = client
	} else {
		return Dispute{}, forbidden
	}
	d, e := scan(tx.QueryRow(ctx, `INSERT INTO disputes(job_id,initiator_id,responder_id,reason,description,evidence_urls) VALUES($1::uuid,$2::uuid,$3::uuid,$4,$5,$6) RETURNING `+cols, job, actor, responder, reason, strings.TrimSpace(desc), ev))
	if e != nil {
		if strings.Contains(e.Error(), "disputes_one_active_job_idx") {
			return d, conflict
		}
		return d, e
	}
	_, e = tx.Exec(ctx, `UPDATE jobs SET status='DISPUTED',updated_at=now() WHERE id=$1::uuid`, job)
	if e == nil {
		_, e = tx.Exec(ctx, `INSERT INTO dispute_audit_log(dispute_id,actor_id,action,to_status) VALUES($1::uuid,$2::uuid,'CREATED','OPEN')`, d.ID, actor)
	}
	if e == nil {
		e = tx.Commit(ctx)
	}
	return d, e
}
func (r *Repository) Get(ctx context.Context, id string) (Dispute, error) {
	d, e := scan(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM disputes WHERE id=$1::uuid`, id))
	if errors.Is(e, pgx.ErrNoRows) {
		e = notFound
	}
	return d, e
}
func authorize(d Dispute, a shared.Actor) error {
	if a.Role == "admin" || a.ID == d.InitiatorID || a.ID == d.ResponderID {
		return nil
	}
	return forbidden
}
func (r *Repository) Evidence(ctx context.Context, id, actor string, urls []string) (Dispute, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Dispute{}, e
	}
	defer tx.Rollback(ctx)
	d, e := scan(tx.QueryRow(ctx, `SELECT `+cols+` FROM disputes WHERE id=$1::uuid FOR UPDATE`, id))
	if errors.Is(e, pgx.ErrNoRows) {
		return d, notFound
	}
	if e != nil {
		return d, e
	}
	if actor != d.InitiatorID && actor != d.ResponderID {
		return d, forbidden
	}
	if d.Status != "OPEN" && d.Status != "UNDER_REVIEW" {
		return d, closed
	}
	d, e = scan(tx.QueryRow(ctx, `UPDATE disputes SET evidence_urls=evidence_urls||$2::text[],updated_at=now() WHERE id=$1::uuid RETURNING `+cols, id, urls))
	if e == nil {
		b, _ := json.Marshal(map[string]any{"evidenceUrls": urls})
		_, e = tx.Exec(ctx, `INSERT INTO dispute_audit_log(dispute_id,actor_id,action,from_status,to_status,details) VALUES($1::uuid,$2::uuid,'EVIDENCE_ADDED',$3,$3,$4)`, id, actor, d.Status, b)
	}
	if e == nil {
		e = tx.Commit(ctx)
	}
	return d, e
}
func (r *Repository) Assign(ctx context.Context, id, admin string) (Dispute, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Dispute{}, e
	}
	defer tx.Rollback(ctx)
	d, e := scan(tx.QueryRow(ctx, `UPDATE disputes SET admin_assigned_id=$2::uuid,status=CASE WHEN status='OPEN' THEN 'UNDER_REVIEW' ELSE status END,updated_at=now() WHERE id=$1::uuid AND status IN ('OPEN','UNDER_REVIEW') RETURNING `+cols, id, admin))
	if errors.Is(e, pgx.ErrNoRows) {
		return d, closed
	}
	if e == nil {
		_, e = tx.Exec(ctx, `INSERT INTO dispute_audit_log(dispute_id,actor_id,action,from_status,to_status,details) VALUES($1::uuid,$2::uuid,'ASSIGNED','OPEN',$3,$4)`, id, admin, d.Status, []byte(`{}`))
	}
	if e == nil {
		e = tx.Commit(ctx)
	}
	return d, e
}
func (r *Repository) Resolve(ctx context.Context, id, admin, status, msg string) (Dispute, error) {
	if status != "RESOLVED" && status != "CLOSED_FAVOR_WORKER" && status != "CLOSED_FAVOR_CLIENT" {
		return Dispute{}, &shared.AppError{Code: "INVALID_STATUS", Status: 400, Message: "Invalid resolution status"}
	}
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Dispute{}, e
	}
	defer tx.Rollback(ctx)
	var old string
	e = tx.QueryRow(ctx, `SELECT status::text FROM disputes WHERE id=$1::uuid AND admin_assigned_id=$2::uuid FOR UPDATE`, id, admin).Scan(&old)
	if errors.Is(e, pgx.ErrNoRows) {
		return Dispute{}, forbidden
	}
	if e != nil {
		return Dispute{}, e
	}
	if old != "UNDER_REVIEW" {
		return Dispute{}, closed
	}
	d, e := scan(tx.QueryRow(ctx, `UPDATE disputes SET status=$2,resolution_message=$3,resolved_at=now(),updated_at=now() WHERE id=$1::uuid RETURNING `+cols, id, status, strings.TrimSpace(msg)))
	if e == nil {
		_, e = tx.Exec(ctx, `INSERT INTO dispute_audit_log(dispute_id,actor_id,action,from_status,to_status,details) VALUES($1::uuid,$2::uuid,'RESOLVED',$3,$4,jsonb_build_object('resolutionMessage',$5))`, id, admin, old, status, msg)
	}
	if e == nil {
		e = tx.Commit(ctx)
	}
	return d, e
}

type PaymentGate interface {
	ValidateDispute(context.Context, string, string) (string, error)
	MarkDisputed(context.Context, string, string) error
}
type Handler struct {
	r        *Repository
	payments PaymentGate
}

func NewHandler(r *Repository) *Handler         { return &Handler{r: r} }
func (h *Handler) SetPaymentGate(g PaymentGate) { h.payments = g }
func (h *Handler) RegisterProtected(m *http.ServeMux, p func(http.Handler) http.Handler) {
	reg := func(s string, f http.HandlerFunc) { m.Handle(s, p(f)) }
	reg("POST /api/v1/disputes", h.create)
	reg("GET /api/v1/disputes/{id}", h.get)
	reg("POST /api/v1/disputes/{id}/evidence", h.evidence)
	reg("PUT /api/v1/admin/disputes/{id}/assign", h.assign)
	reg("PUT /api/v1/admin/disputes/{id}/resolve", h.resolve)
	reg("GET /api/v1/disputes", h.list)
	reg("GET /api/v1/admin/disputes", h.list)
	if h.payments != nil {
		reg("POST /api/v1/payment/{paymentId}/dispute", h.createForPayment)
	}
}
func (h *Handler) createForPayment(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	if a.Role != "client" && a.Role != "worker" {
		fail(w, r, forbidden)
		return
	}
	var in struct {
		Reason       string   `json:"reason"`
		Description  string   `json:"description"`
		EvidenceURLs []string `json:"evidenceUrls"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in) != nil || strings.TrimSpace(in.Reason) == "" {
		shared.WriteLegacyJSON(w, 400, "Dispute reason is required", nil)
		return
	}
	description := strings.TrimSpace(in.Description)
	if description == "" {
		description = strings.TrimSpace(in.Reason)
	}
	jobID, e := h.payments.ValidateDispute(r.Context(), r.PathValue("paymentId"), a.ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	d, e := h.r.Create(r.Context(), a.ID, jobID, "OTHER", description, in.EvidenceURLs)
	if e == nil {
		e = h.payments.MarkDisputed(r.Context(), r.PathValue("paymentId"), a.ID)
	}
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Dispute raised successfully. Our support team will review it and get back to you.", d)
}
func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (shared.Actor, bool) {
	a, ok := shared.ActorFromContext(r.Context())
	if !ok {
		shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), &shared.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Authentication required"})
	}
	return a, ok
}
func fail(w http.ResponseWriter, r *http.Request, e error) {
	shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), e)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	if a.Role != "client" && a.Role != "worker" {
		fail(w, r, forbidden)
		return
	}
	var in struct {
		JobID, Reason, Description string
		EvidenceURLs               []string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.JobID == "" || in.Description == "" {
		shared.WriteLegacyJSON(w, 400, "jobId, reason and description are required", nil)
		return
	}
	d, e := h.r.Create(r.Context(), a.ID, in.JobID, in.Reason, in.Description, in.EvidenceURLs)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Dispute filed successfully", d)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	d, e := h.r.Get(r.Context(), r.PathValue("id"))
	if e == nil {
		e = authorize(d, a)
	}
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Dispute details retrieved successfully", d)
}
func (h *Handler) evidence(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	var in struct {
		EvidenceURLs []string `json:"evidenceUrls"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || len(in.EvidenceURLs) == 0 {
		shared.WriteLegacyJSON(w, 400, "evidenceUrls must be a non-empty array of strings", nil)
		return
	}
	d, e := h.r.Evidence(r.Context(), r.PathValue("id"), a.ID, in.EvidenceURLs)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Evidence added successfully", d)
}
func (h *Handler) assign(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	if a.Role != "admin" {
		fail(w, r, forbidden)
		return
	}
	d, e := h.r.Assign(r.Context(), r.PathValue("id"), a.ID)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Admin assigned to dispute successfully", d)
}
func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	if a.Role != "admin" {
		fail(w, r, forbidden)
		return
	}
	var in struct {
		Status            string `json:"status"`
		ResolutionMessage string `json:"resolutionMessage"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.ResolutionMessage) == "" {
		shared.WriteLegacyJSON(w, 400, "status and resolutionMessage are required", nil)
		return
	}
	d, e := h.r.Resolve(r.Context(), r.PathValue("id"), a.ID, in.Status, in.ResolutionMessage)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Dispute resolved successfully", d)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	a, ok := h.actor(w, r)
	if !ok {
		return
	}
	page, limit := 1, 10
	var e error
	if s := r.URL.Query().Get("page"); s != "" {
		page, e = strconv.Atoi(s)
	}
	if s := r.URL.Query().Get("limit"); s != "" {
		limit, e = strconv.Atoi(s)
	}
	if e != nil || page < 1 || limit < 1 || limit > 100 {
		shared.WriteLegacyJSON(w, 400, "page must be positive and limit between 1 and 100", nil)
		return
	}
	args := []any{}
	where := "TRUE"
	if a.Role != "admin" {
		args = append(args, a.ID)
		where = "(initiator_id=$1::uuid OR responder_id=$1::uuid)"
	}
	if s := r.URL.Query().Get("status"); s != "" {
		args = append(args, s)
		where += " AND status=$" + strconv.Itoa(len(args))
	}
	var p Page
	p.Page, p.Limit = page, limit
	if e = h.r.pool.QueryRow(r.Context(), `SELECT count(*) FROM disputes WHERE `+where, args...).Scan(&p.TotalDocs); e != nil {
		fail(w, r, e)
		return
	}
	args = append(args, limit, (page-1)*limit)
	rows, e := h.r.pool.Query(r.Context(), `SELECT `+cols+` FROM disputes WHERE `+where+` ORDER BY created_at DESC,id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if e != nil {
		fail(w, r, e)
		return
	}
	defer rows.Close()
	p.Docs = []Dispute{}
	for rows.Next() {
		d, x := scan(rows)
		if x != nil {
			fail(w, r, x)
			return
		}
		p.Docs = append(p.Docs, d)
	}
	p.TotalPages = int((p.TotalDocs + int64(limit) - 1) / int64(limit))
	p.PagingCounter = (page-1)*limit + 1
	p.HasPrevPage = page > 1
	p.HasNextPage = page < p.TotalPages
	if p.HasPrevPage {
		x := page - 1
		p.PrevPage = &x
	}
	if p.HasNextPage {
		x := page + 1
		p.NextPage = &x
	}
	shared.WriteLegacyJSON(w, 200, "Disputes retrieved successfully", p)
}
