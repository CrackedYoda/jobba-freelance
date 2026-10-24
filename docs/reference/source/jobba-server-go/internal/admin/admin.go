package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobba/backend/internal/middleware"
	"github.com/jobba/backend/internal/shared"
)

type Role struct {
	ID          string          `json:"_id"`
	Name        string          `json:"name"`
	Permissions json.RawMessage `json:"permissions"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}
type User struct {
	ID          string    `json:"_id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phoneNumber"`
	AccountType string    `json:"accountType"`
	Status      string    `json:"status"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
type Page struct {
	Users       []User `json:"users"`
	TotalDocs   int64  `json:"totalDocs"`
	TotalPages  int    `json:"totalPages"`
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	HasNextPage bool   `json:"hasNextPage"`
}
type UserFilter struct {
	Page, Limit                 int
	AccountType, Status, Search string
}

var errInvalid = &shared.AppError{Code: "INVALID_ADMIN_REQUEST", Status: 400, Message: "Invalid request"}
var errNotFound = &shared.AppError{Code: "ADMIN_RESOURCE_NOT_FOUND", Status: 404, Message: "Resource not found"}
var errConflict = &shared.AppError{Code: "ADMIN_RESOURCE_CONFLICT", Status: 409, Message: "Resource is in use or already exists"}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }
func (r *Repository) audit(ctx context.Context, tx pgx.Tx, actor, action, targetType, target string, metadata any) error {
	value, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO admin_audit_log(actor_id,action,target_type,target_id,metadata) VALUES($1::uuid,$2,$3,$4::uuid,$5::jsonb)`, actor, action, targetType, target, value)
	return err
}
func scanRole(row pgx.Row) (Role, error) {
	var v Role
	err := row.Scan(&v.ID, &v.Name, &v.Permissions, &v.CreatedAt, &v.UpdatedAt)
	return v, err
}
func (r *Repository) Roles(ctx context.Context) ([]Role, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,name,permissions,created_at,updated_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Role{}
	for rows.Next() {
		v, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *Repository) Role(ctx context.Context, id string) (Role, error) {
	v, err := scanRole(r.pool.QueryRow(ctx, `SELECT id::text,name,permissions,created_at,updated_at FROM roles WHERE id=$1::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return v, errNotFound
	}
	return v, err
}
func (r *Repository) CreateRole(ctx context.Context, actor, name string, permissions json.RawMessage) (Role, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Role{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	v, err := scanRole(tx.QueryRow(ctx, `INSERT INTO roles(name,permissions) VALUES(lower($1),$2::jsonb) RETURNING id::text,name,permissions,created_at,updated_at`, name, permissions))
	if err != nil {
		return Role{}, errConflict
	}
	if err = r.audit(ctx, tx, actor, "role.create", "role", v.ID, map[string]any{"name": v.Name}); err != nil {
		return Role{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Role{}, err
	}
	return v, nil
}
func (r *Repository) UpdateRole(ctx context.Context, actor, id, name string, permissions json.RawMessage) (Role, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Role{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	v, err := scanRole(tx.QueryRow(ctx, `UPDATE roles SET name=lower($2),permissions=$3::jsonb,updated_at=now() WHERE id=$1::uuid RETURNING id::text,name,permissions,created_at,updated_at`, id, name, permissions))
	if errors.Is(err, pgx.ErrNoRows) {
		return Role{}, errNotFound
	}
	if err != nil {
		return Role{}, errConflict
	}
	if err = r.audit(ctx, tx, actor, "role.update", "role", id, map[string]any{"name": v.Name}); err != nil {
		return Role{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Role{}, err
	}
	return v, nil
}
func (r *Repository) DeleteRole(ctx context.Context, actor, id string) (Role, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Role{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var count int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE role_id=$1::uuid`, id).Scan(&count); err != nil {
		return Role{}, err
	}
	if count > 0 {
		return Role{}, errConflict
	}
	v, err := scanRole(tx.QueryRow(ctx, `DELETE FROM roles WHERE id=$1::uuid RETURNING id::text,name,permissions,created_at,updated_at`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Role{}, errNotFound
	}
	if err != nil {
		return Role{}, err
	}
	if err = r.audit(ctx, tx, actor, "role.delete", "role", id, map[string]any{"name": v.Name}); err != nil {
		return Role{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Role{}, err
	}
	return v, nil
}
func (r *Repository) Users(ctx context.Context, f UserFilter) (Page, error) {
	args := []any{}
	where := []string{"true"}
	add := func(value any) string { args = append(args, value); return fmt.Sprintf("$%d", len(args)) }
	if f.AccountType != "" {
		where = append(where, "account_type="+add(f.AccountType))
	}
	if f.Status != "" {
		where = append(where, "status="+add(f.Status))
	}
	if f.Search != "" {
		p := add("%" + f.Search + "%")
		where = append(where, "(first_name ILIKE "+p+" OR last_name ILIKE "+p+" OR email ILIKE "+p+" OR phone_number ILIKE "+p+")")
	}
	clause := strings.Join(where, " AND ")
	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE "+clause, args...).Scan(&total); err != nil {
		return Page{}, err
	}
	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	rows, err := r.pool.Query(ctx, "SELECT id::text,first_name,last_name,email,phone_number,account_type,status,is_active,created_at,updated_at FROM users WHERE "+clause+fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PhoneNumber, &u.AccountType, &u.Status, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return Page{}, err
		}
		users = append(users, u)
	}
	pages := int((total + int64(f.Limit) - 1) / int64(f.Limit))
	return Page{Users: users, TotalDocs: total, TotalPages: pages, Page: f.Page, Limit: f.Limit, HasNextPage: f.Page < pages}, rows.Err()
}
func (r *Repository) UpdateUser(ctx context.Context, actor, id, field, value string) (User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if actor == id {
		return User{}, errConflict
	}
	query := ""
	switch field {
	case "status":
		query = `UPDATE users u SET status=$2,is_active=($2='ACTIVE'),updated_at=now() WHERE id=$1::uuid`
	case "account_type":
		query = `UPDATE users u SET account_type=$2,role_id=r.id,updated_at=now() FROM roles r WHERE u.id=$1::uuid AND r.name=$2`
	default:
		return User{}, errInvalid
	}
	query += ` RETURNING u.id::text,u.first_name,u.last_name,u.email,u.phone_number,u.account_type,u.status,u.is_active,u.created_at,u.updated_at`
	var u User
	err = tx.QueryRow(ctx, query, id, value).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PhoneNumber, &u.AccountType, &u.Status, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, errNotFound
	}
	if err != nil {
		return User{}, err
	}
	if err = r.audit(ctx, tx, actor, "user."+field, "user", id, map[string]any{"value": value}); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return u, nil
}

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }
func (h *Handler) RegisterProtected(mux *http.ServeMux, protect func(http.Handler) http.Handler) {
	admin := func(next http.HandlerFunc) http.Handler {
		return protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a, ok := shared.ActorFromContext(r.Context())
			if !ok || a.Role != "admin" {
				shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), &shared.AppError{Code: "FORBIDDEN", Status: 403, Message: "You don't have permission to perform this action"})
				return
			}
			next(w, r)
		}))
	}
	mux.Handle("GET /api/v1/admin/users", admin(h.users))
	mux.Handle("PATCH /api/v1/admin/users/{id}/status", admin(h.userStatus))
	mux.Handle("PATCH /api/v1/admin/users/{id}/account-type", admin(h.userAccount))
	mux.Handle("GET /api/v1/role/all", admin(h.roles))
	mux.Handle("GET /api/v1/role/single/{id}", admin(h.role))
	mux.Handle("POST /api/v1/role/new", admin(h.createRole))
	mux.Handle("PUT /api/v1/role/update/{id}", admin(h.updateRole))
	mux.Handle("DELETE /api/v1/role/delete/{id}", admin(h.deleteRole))
}
func actor(r *http.Request) string { a, _ := shared.ActorFromContext(r.Context()); return a.ID }
func fail(w http.ResponseWriter, r *http.Request, e error) {
	shared.WriteLegacyError(w, middleware.LoggerFromContext(r.Context()), e)
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func (h *Handler) users(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	f := UserFilter{Page: page, Limit: limit, AccountType: q.Get("accountType"), Status: q.Get("status"), Search: strings.TrimSpace(q.Get("search"))}
	if f.AccountType != "" && f.AccountType != "client" && f.AccountType != "worker" && f.AccountType != "admin" {
		fail(w, r, errInvalid)
		return
	}
	v, e := h.repo.Users(r.Context(), f)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Users fetched successfully", v)
}
func (h *Handler) roles(w http.ResponseWriter, r *http.Request) {
	v, e := h.repo.Roles(r.Context())
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Roles retrieved", v)
}
func (h *Handler) role(w http.ResponseWriter, r *http.Request) {
	v, e := h.repo.Role(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Role retrieved", v)
}

type roleBody struct {
	Name        string          `json:"name"`
	Permissions json.RawMessage `json:"permissions"`
}

func validRole(v roleBody) bool {
	if len(strings.TrimSpace(v.Name)) < 2 || !json.Valid(v.Permissions) {
		return false
	}
	var object map[string]any
	return json.Unmarshal(v.Permissions, &object) == nil && object["resources"] != nil
}
func (h *Handler) createRole(w http.ResponseWriter, r *http.Request) {
	var b roleBody
	if decode(w, r, &b) != nil || !validRole(b) {
		fail(w, r, errInvalid)
		return
	}
	v, e := h.repo.CreateRole(r.Context(), actor(r), b.Name, b.Permissions)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Role created", v)
}
func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request) {
	var b roleBody
	if decode(w, r, &b) != nil || !validRole(b) {
		fail(w, r, errInvalid)
		return
	}
	v, e := h.repo.UpdateRole(r.Context(), actor(r), r.PathValue("id"), b.Name, b.Permissions)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "Role updated!", v)
}
func (h *Handler) deleteRole(w http.ResponseWriter, r *http.Request) {
	v, e := h.repo.DeleteRole(r.Context(), actor(r), r.PathValue("id"))
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 201, "Role deleted!", v)
}
func (h *Handler) userStatus(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Status string `json:"status"`
	}
	if decode(w, r, &b) != nil || (b.Status != "ACTIVE" && b.Status != "INACTIVE" && b.Status != "SUSPENDED" && b.Status != "DELETED" && b.Status != "DEACTIVATED") {
		fail(w, r, errInvalid)
		return
	}
	v, e := h.repo.UpdateUser(r.Context(), actor(r), r.PathValue("id"), "status", b.Status)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "User status updated", v)
}
func (h *Handler) userAccount(w http.ResponseWriter, r *http.Request) {
	var b struct {
		AccountType string `json:"accountType"`
	}
	if decode(w, r, &b) != nil || (b.AccountType != "client" && b.AccountType != "worker" && b.AccountType != "admin") {
		fail(w, r, errInvalid)
		return
	}
	v, e := h.repo.UpdateUser(r.Context(), actor(r), r.PathValue("id"), "account_type", b.AccountType)
	if e != nil {
		fail(w, r, e)
		return
	}
	shared.WriteLegacyJSON(w, 200, "User account type updated", v)
}
