package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &repository{pool: pool} }

type scanner interface{ Scan(...any) error }

const applicationColumns = `id::text, job_id::text, worker_id::text, cover_letter, bid_amount::float8, work_duration, status::text, attachments, created_at, updated_at`

func scanApplication(row scanner) (Application, error) {
	var a Application
	var status string
	err := row.Scan(&a.ID, &a.JobID, &a.WorkerID, &a.CoverLetter, &a.BidAmount, &a.WorkDuration, &status, &a.Attachments, &a.CreatedAt, &a.UpdatedAt)
	a.Status, a.Job, a.Worker = Status(status), a.JobID, a.WorkerID
	return a, err
}

func (r *repository) GetJob(ctx context.Context, id string) (JobState, error) {
	var job JobState
	err := r.pool.QueryRow(ctx, `SELECT id::text,client_id::text,status::text,is_deleted,privacy::text,applicant_type::text FROM jobs WHERE id=$1::uuid`, id).Scan(&job.ID, &job.ClientID, &job.Status, &job.IsDeleted, &job.Privacy, &job.ApplicantType)
	if errors.Is(err, pgx.ErrNoRows) {
		return JobState{}, ErrJobNotFound
	}
	return job, err
}
func (r *repository) HasInvitation(ctx context.Context, jobID, workerID string) (bool, error) {
	var invited bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM offers WHERE job_id=$1::uuid AND worker_id=$2::uuid AND status='PENDING')`, jobID, workerID).Scan(&invited)
	return invited, err
}

func (r *repository) Create(ctx context.Context, jobID, workerID string, in CreateInput) (Application, error) {
	q := `INSERT INTO applications(job_id,worker_id,cover_letter,bid_amount,work_duration,attachments) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6) RETURNING ` + applicationColumns
	a, err := scanApplication(r.pool.QueryRow(ctx, q, jobID, workerID, in.CoverLetter, in.BidAmount, in.WorkDuration, in.Attachments))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Application{}, ErrAlreadyApplied
	}
	return a, err
}

func (r *repository) ListWorker(ctx context.Context, id string, in ListInput) (Page, error) {
	return r.list(ctx, `a.worker_id=$1::uuid`, id, in, true)
}

func (r *repository) ListJob(ctx context.Context, id string, in ListInput) (Page, error) {
	return r.list(ctx, `a.job_id=$1::uuid`, id, in, false)
}

func (r *repository) list(ctx context.Context, where, id string, in ListInput, populateJob bool) (Page, error) {
	args := []any{id}
	if in.Status != nil {
		args = append(args, *in.Status)
		where += fmt.Sprintf(` AND a.status=$%d`, len(args))
	}
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM applications a WHERE `+where, args...).Scan(&total); err != nil {
		return Page{}, err
	}
	args = append(args, in.Limit, (in.Page-1)*in.Limit)
	rows, err := r.pool.Query(ctx, `SELECT `+applicationColumns+` FROM applications a WHERE `+where+fmt.Sprintf(` ORDER BY a.created_at DESC,a.id DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	page := newPage(total, in)
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return Page{}, err
		}
		if populateJob {
			err = r.populateJob(ctx, &a)
		} else {
			err = r.populateWorker(ctx, &a)
		}
		if err != nil {
			return Page{}, err
		}
		page.Docs = append(page.Docs, a)
	}
	return page, rows.Err()
}

func newPage(total int64, in ListInput) Page {
	p := Page{Docs: []Application{}, TotalDocs: total, Limit: in.Limit, Page: in.Page, PagingCounter: (in.Page-1)*in.Limit + 1}
	p.TotalPages = int((total + int64(in.Limit) - 1) / int64(in.Limit))
	p.HasPrevPage = in.Page > 1 && p.TotalPages > 0
	p.HasNextPage = in.Page < p.TotalPages
	if p.HasPrevPage {
		v := in.Page - 1
		p.PrevPage = &v
	}
	if p.HasNextPage {
		v := in.Page + 1
		p.NextPage = &v
	}
	return p
}

func (r *repository) populateWorker(ctx context.Context, a *Application) error {
	var u UserSummary
	err := r.pool.QueryRow(ctx, `SELECT id::text,first_name,last_name,email,phone_number FROM users WHERE id=$1::uuid`, a.WorkerID).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PhoneNumber)
	if err == nil {
		a.Worker = u
	}
	return err
}

func (r *repository) populateJob(ctx context.Context, a *Application) error {
	var j JobSummary
	err := r.pool.QueryRow(ctx, `SELECT j.id::text,j.job_title,j.min_budget::float8,j.max_budget::float8,j.work_type::text,j.status::text,u.id::text,u.first_name,u.last_name FROM jobs j JOIN users u ON u.id=j.client_id WHERE j.id=$1::uuid`, a.JobID).Scan(&j.ID, &j.JobTitle, &j.MinBudget, &j.MaxBudget, &j.WorkType, &j.Status, &j.Client.ID, &j.Client.FirstName, &j.Client.LastName)
	if err == nil {
		a.Job = j
	}
	return err
}

func (r *repository) Accept(ctx context.Context, applicationID, clientID string) (Application, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Application{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	a, err := scanApplication(tx.QueryRow(ctx, `SELECT `+applicationColumns+` FROM applications a WHERE a.id=$1::uuid FOR UPDATE`, applicationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrApplicationNotFound
	}
	if err != nil {
		return Application{}, err
	}
	var owner, status string
	var deleted bool
	err = tx.QueryRow(ctx, `SELECT client_id::text,status::text,is_deleted FROM jobs WHERE id=$1::uuid FOR UPDATE`, a.JobID).Scan(&owner, &status, &deleted)
	if err != nil {
		return Application{}, err
	}
	if owner != clientID {
		return Application{}, ErrAcceptForbidden
	}
	if deleted || status != "OPEN" {
		return Application{}, ErrJobInactive
	}
	if _, err = tx.Exec(ctx, `UPDATE applications SET status='ACCEPTED',updated_at=now() WHERE id=$1::uuid`, a.ID); err != nil {
		return Application{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE applications SET status='REJECTED',updated_at=now() WHERE job_id=$1::uuid AND id<>$2::uuid AND status IN ('PENDING','SHORTLISTED')`, a.JobID, a.ID); err != nil {
		return Application{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE jobs SET assigned_worker_id=$2::uuid,status='ACTIVE',updated_at=now() WHERE id=$1::uuid`, a.JobID, a.WorkerID); err != nil {
		return Application{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Application{}, err
	}
	a.Status = StatusAccepted
	if err = r.populateJob(ctx, &a); err != nil {
		return Application{}, err
	}
	return a, nil
}
