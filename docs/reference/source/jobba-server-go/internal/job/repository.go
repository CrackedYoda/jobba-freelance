package job

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &repository{pool: pool} }

const jobColumns = `id::text, COALESCE(legacy_id, id::text), client_id::text, job_title, category, description, attachments, min_budget::float8, max_budget::float8, payment_method::text, start_date, end_date, work_type::text, location, applicant_type::text, privacy::text, status::text, assigned_worker_id::text, budget::float8, view_count, created_at, updated_at`

type rowScanner interface{ Scan(...any) error }

func scanJob(row rowScanner) (Job, error) {
	var j Job
	var payment, work, applicant, privacy, status string
	err := row.Scan(&j.ID, &j.UniqueID, &j.ClientID, &j.JobTitle, &j.Category, &j.Description, &j.Attachments, &j.MinBudget, &j.MaxBudget, &payment, &j.StartDate, &j.EndDate, &work, &j.Location, &applicant, &privacy, &status, &j.AssignedWorkerID, &j.Budget, &j.ViewCount, &j.CreatedAt, &j.UpdatedAt)
	j.PaymentMethod = PaymentMethod(payment)
	j.WorkType = WorkType(work)
	j.ApplicantType = ApplicantType(applicant)
	j.Privacy = Privacy(privacy)
	j.Status = Status(status)
	j.Skills = []Skill{}
	return j, err
}
func (r *repository) skills(ctx context.Context, j *Job) error {
	rows, e := r.pool.Query(ctx, `SELECT s.id::text,s.name FROM skills s JOIN job_skills js ON js.skill_id=s.id WHERE js.job_id=$1 ORDER BY s.name`, j.ID)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var s Skill
		if e = rows.Scan(&s.ID, &s.Name); e != nil {
			return e
		}
		j.Skills = append(j.Skills, s)
	}
	return rows.Err()
}
func (r *repository) Create(ctx context.Context, client string, in CreateInput) (Job, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Job{}, e
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := `INSERT INTO jobs(client_id,job_title,category,description,attachments,min_budget,max_budget,payment_method,start_date,end_date,work_type,location,applicant_type,privacy) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING ` + jobColumns
	j, e := scanJob(tx.QueryRow(ctx, q, client, in.JobTitle, in.Category, in.Description, in.Attachments, in.MinBudget, in.MaxBudget, in.PaymentMethod, in.StartDate, in.EndDate, in.WorkType, in.Location, in.ApplicantType, in.Privacy))
	if e != nil {
		return Job{}, e
	}
	for _, skill := range in.SkillIDs {
		if _, e = tx.Exec(ctx, `INSERT INTO job_skills(job_id,skill_id) VALUES($1,$2::uuid)`, j.ID, skill); e != nil {
			return Job{}, e
		}
	}
	if e = tx.Commit(ctx); e != nil {
		return Job{}, e
	}
	return j, r.skills(ctx, &j)
}
func (r *repository) GetOwned(ctx context.Context, id, client string) (Job, error) {
	j, e := scanJob(r.pool.QueryRow(ctx, `SELECT `+jobColumns+` FROM jobs WHERE id=$1::uuid AND client_id=$2::uuid AND is_deleted=false`, id, client))
	if errors.Is(e, pgx.ErrNoRows) {
		return Job{}, ErrNotFound
	}
	if e == nil {
		e = r.skills(ctx, &j)
	}
	return j, e
}
func (r *repository) GetVisible(ctx context.Context, id string) (Job, error) {
	j, e := scanJob(r.pool.QueryRow(ctx, `SELECT `+jobColumns+` FROM jobs WHERE id=$1::uuid AND is_deleted=false`, id))
	if errors.Is(e, pgx.ErrNoRows) {
		return Job{}, ErrNotFound
	}
	if e == nil {
		e = r.skills(ctx, &j)
	}
	return j, e
}
func (r *repository) WorkerCanView(ctx context.Context, jobID, workerID string) (bool, error) {
	var allowed bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM offers WHERE job_id=$1::uuid AND worker_id=$2::uuid AND status IN ('PENDING','ACCEPTED'))`, jobID, workerID).Scan(&allowed)
	return allowed, err
}
func (r *repository) List(ctx context.Context, f ListFilter) (Page, error) {
	where := []string{"is_deleted=false"}
	args := []any{}
	add := func(s string, v any) { args = append(args, v); where = append(where, fmt.Sprintf(s, len(args))) }
	if f.ClientID != "" {
		add("client_id=$%d::uuid", f.ClientID)
	}
	if f.WorkerID != "" {
		add("assigned_worker_id=$%d::uuid", f.WorkerID)
	}
	if f.Category != "" {
		add("category=$%d", f.Category)
	}
	if f.Status != nil {
		add("status=$%d", *f.Status)
	}
	if f.WorkType != nil {
		add("work_type=$%d", *f.WorkType)
	}
	if f.MinBudget != nil {
		add("min_budget >= $%d", *f.MinBudget)
	}
	if f.MaxBudget != nil {
		add("max_budget <= $%d", *f.MaxBudget)
	}
	if f.Search != "" {
		args = append(args, f.Search)
		where = append(where, fmt.Sprintf("(job_title ILIKE '%%'||$%d||'%%' OR description ILIKE '%%'||$%d||'%%')", len(args), len(args)))
	}
	if f.PublicOnly {
		where = append(where, "privacy='public'")
	}
	if f.UnassignedOnly {
		where = append(where, "assigned_worker_id IS NULL")
	}
	if len(f.SkillIDs) > 0 {
		args = append(args, f.SkillIDs)
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM job_skills js WHERE js.job_id=jobs.id AND js.skill_id=ANY($%d::uuid[]))", len(args)))
	}
	sorts := map[string]string{"createdAt": "created_at", "minBudget": "min_budget", "maxBudget": "max_budget", "startDate": "start_date"}
	order := "DESC"
	if f.Order == "asc" {
		order = "ASC"
	}
	base := " FROM jobs WHERE " + strings.Join(where, " AND ")
	var total int64
	if e := r.pool.QueryRow(ctx, "SELECT count(*)"+base, args...).Scan(&total); e != nil {
		return Page{}, e
	}
	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	q := "SELECT " + jobColumns + base + fmt.Sprintf(" ORDER BY %s %s,id ASC LIMIT $%d OFFSET $%d", sorts[f.SortBy], order, len(args)-1, len(args))
	rows, e := r.pool.Query(ctx, q, args...)
	if e != nil {
		return Page{}, e
	}
	defer rows.Close()
	p := Page{TotalCount: total, Data: []Job{}}
	for rows.Next() {
		j, e := scanJob(rows)
		if e != nil {
			return Page{}, e
		}
		if e = r.skills(ctx, &j); e != nil {
			return Page{}, e
		}
		p.Data = append(p.Data, j)
	}
	if e = rows.Err(); e != nil {
		return Page{}, e
	}
	p.Count = len(p.Data)
	p.Pages = int((total + int64(f.Limit) - 1) / int64(f.Limit))
	if f.Page < p.Pages {
		n := f.Page + 1
		p.Next = &n
	}
	if f.Page > 1 && p.Pages > 0 {
		v := f.Page - 1
		p.Prev = &v
	}
	return p, nil
}
func (r *repository) Update(ctx context.Context, current Job, in UpdateInput) (Job, error) {
	if in.JobTitle != nil {
		current.JobTitle = strings.TrimSpace(*in.JobTitle)
	}
	if in.Category != nil {
		current.Category = strings.TrimSpace(*in.Category)
	}
	if in.Description != nil {
		current.Description = strings.TrimSpace(*in.Description)
	}
	if in.Attachments != nil {
		current.Attachments = *in.Attachments
	}
	if in.MinBudget != nil {
		current.MinBudget = *in.MinBudget
	}
	if in.MaxBudget != nil {
		current.MaxBudget = *in.MaxBudget
	}
	if in.PaymentMethod != nil {
		current.PaymentMethod = *in.PaymentMethod
	}
	if in.StartDate != nil {
		current.StartDate = *in.StartDate
	}
	if in.EndDate != nil {
		current.EndDate = *in.EndDate
	}
	if in.WorkType != nil {
		current.WorkType = *in.WorkType
	}
	if in.Location != nil {
		current.Location = strings.TrimSpace(*in.Location)
	}
	if in.ApplicantType != nil {
		current.ApplicantType = *in.ApplicantType
	}
	if in.Privacy != nil {
		current.Privacy = *in.Privacy
	}
	if in.Status != nil {
		current.Status = *in.Status
	}
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Job{}, e
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := `UPDATE jobs SET job_title=$2,category=$3,description=$4,attachments=$5,min_budget=$6,max_budget=$7,payment_method=$8,start_date=$9,end_date=$10,work_type=$11,location=$12,applicant_type=$13,privacy=$14,status=$15,updated_at=now() WHERE id=$1::uuid RETURNING ` + jobColumns
	j, e := scanJob(tx.QueryRow(ctx, q, current.ID, current.JobTitle, current.Category, current.Description, current.Attachments, current.MinBudget, current.MaxBudget, current.PaymentMethod, current.StartDate, current.EndDate, current.WorkType, current.Location, current.ApplicantType, current.Privacy, current.Status))
	if e != nil {
		return Job{}, e
	}
	if in.SkillIDs != nil {
		if _, e = tx.Exec(ctx, `DELETE FROM job_skills WHERE job_id=$1::uuid`, j.ID); e != nil {
			return Job{}, e
		}
		for _, skill := range *in.SkillIDs {
			if _, e = tx.Exec(ctx, `INSERT INTO job_skills(job_id,skill_id) VALUES($1::uuid,$2::uuid)`, j.ID, skill); e != nil {
				return Job{}, e
			}
		}
	}
	if e = tx.Commit(ctx); e != nil {
		return Job{}, e
	}
	return j, r.skills(ctx, &j)
}
func (r *repository) SoftDelete(ctx context.Context, id string) error {
	tag, e := r.pool.Exec(ctx, `UPDATE jobs SET is_deleted=true,status='CLOSED',updated_at=now() WHERE id=$1::uuid AND is_deleted=false`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return e
}
func (r *repository) SetStatus(ctx context.Context, id string, status Status) (Job, error) {
	j, e := scanJob(r.pool.QueryRow(ctx, `UPDATE jobs SET status=$2,updated_at=now() WHERE id=$1::uuid RETURNING `+jobColumns, id, status))
	if e == nil {
		e = r.skills(ctx, &j)
	}
	return j, e
}
func (r *repository) Assign(ctx context.Context, id, worker string) (Job, error) {
	j, e := scanJob(r.pool.QueryRow(ctx, `UPDATE jobs SET assigned_worker_id=$2::uuid,status='ACTIVE',updated_at=now() WHERE id=$1::uuid RETURNING `+jobColumns, id, worker))
	if e == nil {
		e = r.skills(ctx, &j)
	}
	return j, e
}
func (r *repository) WorkerExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1::uuid AND account_type='worker')`, id).Scan(&exists)
	return exists, err
}
func (r *repository) IncrementView(ctx context.Context, id string) (Job, error) {
	j, e := scanJob(r.pool.QueryRow(ctx, `UPDATE jobs SET view_count=view_count+1 WHERE id=$1::uuid RETURNING `+jobColumns, id))
	if e == nil {
		e = r.skills(ctx, &j)
	}
	return j, e
}
func (r *repository) Statistics(ctx context.Context, client string) (Statistics, error) {
	s := Statistics{StatusBreakdown: []StatusCount{}}
	if e := r.pool.QueryRow(ctx, `SELECT count(*),COALESCE(sum(max_budget) FILTER(WHERE status='COMPLETED'),0)::float8 FROM jobs WHERE client_id=$1::uuid AND is_deleted=false`, client).Scan(&s.TotalJobs, &s.TotalSpent); e != nil {
		return s, e
	}
	rows, e := r.pool.Query(ctx, `SELECT status::text,count(*) FROM jobs WHERE client_id=$1::uuid AND is_deleted=false GROUP BY status ORDER BY status`, client)
	if e != nil {
		return s, e
	}
	defer rows.Close()
	for rows.Next() {
		var v StatusCount
		if e = rows.Scan(&v.ID, &v.Count); e != nil {
			return s, e
		}
		s.StatusBreakdown = append(s.StatusBreakdown, v)
	}
	return s, rows.Err()
}
func (r *repository) Save(ctx context.Context, user, id string) error {
	_, e := r.pool.Exec(ctx, `INSERT INTO saved_jobs(user_id,job_id) VALUES($1::uuid,$2::uuid) ON CONFLICT DO NOTHING`, user, id)
	return e
}
func (r *repository) Unsave(ctx context.Context, user, id string) error {
	_, e := r.pool.Exec(ctx, `DELETE FROM saved_jobs WHERE user_id=$1::uuid AND job_id=$2::uuid`, user, id)
	return e
}
func (r *repository) ListSaved(ctx context.Context, user string, p Pagination) (Page, error) {
	var total int64
	if e := r.pool.QueryRow(ctx, `SELECT count(*) FROM saved_jobs sj JOIN jobs j ON j.id=sj.job_id WHERE sj.user_id=$1::uuid AND j.is_deleted=false`, user).Scan(&total); e != nil {
		return Page{}, e
	}
	rows, e := r.pool.Query(ctx, `SELECT `+jobColumns+` FROM jobs WHERE id IN (SELECT job_id FROM saved_jobs WHERE user_id=$1::uuid) AND is_deleted=false ORDER BY (SELECT created_at FROM saved_jobs WHERE user_id=$1::uuid AND job_id=jobs.id) DESC LIMIT $2 OFFSET $3`, user, p.Limit, (p.Page-1)*p.Limit)
	if e != nil {
		return Page{}, e
	}
	defer rows.Close()
	result := Page{TotalCount: total, Data: []Job{}}
	for rows.Next() {
		j, e := scanJob(rows)
		if e != nil {
			return Page{}, e
		}
		if e = r.skills(ctx, &j); e != nil {
			return Page{}, e
		}
		result.Data = append(result.Data, j)
	}
	if e = rows.Err(); e != nil {
		return Page{}, e
	}
	result.Count = len(result.Data)
	result.Pages = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	if p.Page < result.Pages {
		n := p.Page + 1
		result.Next = &n
	}
	if p.Page > 1 && result.Pages > 0 {
		v := p.Page - 1
		result.Prev = &v
	}
	return result, nil
}
