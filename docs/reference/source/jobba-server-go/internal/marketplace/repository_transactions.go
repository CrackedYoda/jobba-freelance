package marketplace

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func scanOffer(row interface{ Scan(...any) error }) (Offer, error) {
	var o Offer
	e := row.Scan(&o.ID, &o.JobID, &o.ClientID, &o.WorkerID, &o.Budget, &o.Title, &o.Description, &o.StartDate, &o.EndDate, &o.Location, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, e
}

const offerCols = `id::text,job_id::text,client_id::text,worker_id::text,budget::float8,COALESCE(title,''),COALESCE(description,''),start_date,end_date,COALESCE(location,''),status::text,created_at,updated_at`

func (r *repository) CreateOffer(c context.Context, client string, in OfferInput) (Offer, error) {
	tx, e := r.pool.Begin(c)
	if e != nil {
		return Offer{}, e
	}
	defer func() { _ = tx.Rollback(c) }()
	var status string
	var assigned *string
	e = tx.QueryRow(c, `SELECT status::text,assigned_worker_id::text FROM jobs WHERE id=$1::uuid AND client_id=$2::uuid AND NOT is_deleted FOR UPDATE`, in.JobID, client).Scan(&status, &assigned)
	if errors.Is(e, pgx.ErrNoRows) {
		return Offer{}, ErrJobNotFound
	}
	if e != nil {
		return Offer{}, e
	}
	if status != "OPEN" || assigned != nil {
		return Offer{}, ErrJobState
	}
	var worker bool
	if e = tx.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1::uuid AND account_type='worker' AND is_active)`, in.WorkerID).Scan(&worker); e != nil {
		return Offer{}, e
	}
	if !worker {
		return Offer{}, ErrWorkerNotFound
	}
	if _, e = tx.Exec(c, `UPDATE offers SET status='REJECTED',updated_at=now() WHERE job_id=$1::uuid AND worker_id=$2::uuid AND status='PENDING'`, in.JobID, in.WorkerID); e != nil {
		return Offer{}, e
	}
	o, e := scanOffer(tx.QueryRow(c, `INSERT INTO offers(job_id,client_id,worker_id,budget,title,description,start_date,end_date,location) VALUES($1::uuid,$2::uuid,$3::uuid,$4,NULLIF($5,''),NULLIF($6,''),$7,$8,NULLIF($9,'')) RETURNING `+offerCols, in.JobID, client, in.WorkerID, in.Budget, in.Title, in.Description, in.StartDate, in.EndDate, in.Location))
	if e != nil {
		return Offer{}, e
	}
	return o, tx.Commit(c)
}
func (r *repository) AcceptOffer(c context.Context, id, worker string) (Job, error) {
	tx, e := r.pool.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return Job{}, e
	}
	defer func() { _ = tx.Rollback(c) }()
	var o Offer
	o, e = scanOffer(tx.QueryRow(c, `SELECT `+offerCols+` FROM offers WHERE id=$1::uuid AND worker_id=$2::uuid AND status='PENDING' FOR UPDATE`, id, worker))
	if errors.Is(e, pgx.ErrNoRows) {
		return Job{}, ErrOfferNotFound
	}
	if e != nil {
		return Job{}, e
	}
	var j Job
	var assigned *string
	e = tx.QueryRow(c, `SELECT id::text,job_title,category,description,min_budget::float8,max_budget::float8,work_type::text,location,status::text,client_id::text,created_at,assigned_worker_id::text FROM jobs WHERE id=$1::uuid AND NOT is_deleted FOR UPDATE`, o.JobID).Scan(&j.ID, &j.JobTitle, &j.Category, &j.Description, &j.MinBudget, &j.MaxBudget, &j.WorkType, &j.Location, &j.Status, &j.ClientID, &j.CreatedAt, &assigned)
	if errors.Is(e, pgx.ErrNoRows) {
		return Job{}, ErrJobNotFound
	}
	if e != nil {
		return Job{}, e
	}
	if assigned != nil || j.Status != "OPEN" {
		return Job{}, ErrJobState
	}
	_, e = tx.Exec(c, `UPDATE jobs SET assigned_worker_id=$2::uuid,budget=$3,status='ACTIVE',job_title=COALESCE(NULLIF($4,''),job_title),description=COALESCE(NULLIF($5,''),description),start_date=COALESCE($6,start_date),end_date=COALESCE($7,end_date),location=COALESCE(NULLIF($8,''),location),updated_at=now() WHERE id=$1::uuid`, j.ID, worker, o.Budget, o.Title, o.Description, o.StartDate, o.EndDate, o.Location)
	if e != nil {
		return Job{}, e
	}
	if _, e = tx.Exec(c, `UPDATE offers SET status='ACCEPTED',updated_at=now() WHERE id=$1::uuid`, id); e != nil {
		return Job{}, e
	}
	if _, e = tx.Exec(c, `UPDATE applications SET status='REJECTED',updated_at=now() WHERE job_id=$1::uuid AND status='PENDING'`, j.ID); e != nil {
		return Job{}, e
	}
	if e = tx.Commit(c); e != nil {
		return Job{}, e
	}
	j.Status = "ACTIVE"
	return j, nil
}
func (r *repository) RejectOffer(c context.Context, id, worker string) error {
	tag, e := r.pool.Exec(c, `UPDATE offers SET status='REJECTED',updated_at=now() WHERE id=$1::uuid AND worker_id=$2::uuid AND status='PENDING'`, id, worker)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrOfferNotFound
	}
	return e
}
func (r *repository) CreateReview(c context.Context, user, jobID string, rating int, comment string) (Review, error) {
	var client string
	var worker *string
	var status string
	e := r.pool.QueryRow(c, `SELECT client_id::text,assigned_worker_id::text,status::text FROM jobs WHERE id=$1::uuid`, jobID).Scan(&client, &worker, &status)
	if errors.Is(e, pgx.ErrNoRows) {
		return Review{}, ErrJobNotFound
	}
	if e != nil {
		return Review{}, e
	}
	if status != "COMPLETED" {
		return Review{}, ErrJobState
	}
	reviewee := ""
	if client == user && worker != nil {
		reviewee = *worker
	} else if worker != nil && *worker == user {
		reviewee = client
	} else {
		return Review{}, ErrReviewForbidden
	}
	var v Review
	var jid, rid string
	e = r.pool.QueryRow(c, `INSERT INTO reviews(job_id,reviewer_id,reviewee_id,rating,comment) VALUES($1::uuid,$2::uuid,$3::uuid,$4,NULLIF($5,'')) RETURNING id::text,job_id::text,reviewer_id::text,reviewee_id::text,rating,COALESCE(comment,''),created_at,updated_at`, jobID, user, reviewee, rating, comment).Scan(&v.ID, &jid, &rid, &v.Reviewee, &v.Rating, &v.Comment, &v.CreatedAt, &v.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(e, &pgErr) && pgErr.Code == "23505" {
		return Review{}, ErrAlreadyReviewed
	}
	v.Job = jid
	v.Reviewer = rid
	return v, e
}
func (r *repository) Reviews(c context.Context, id string) (ReviewResult, error) {
	rows, e := r.pool.Query(c, `SELECT r.id::text,j.id::text,j.job_title,u.id::text,u.first_name,u.last_name,u.profile_image,r.rating,COALESCE(r.comment,''),r.created_at,r.updated_at FROM reviews r JOIN jobs j ON j.id=r.job_id JOIN users u ON u.id=r.reviewer_id WHERE r.reviewee_id=$1::uuid ORDER BY r.created_at DESC`, id)
	if e != nil {
		return ReviewResult{}, e
	}
	defer rows.Close()
	out := ReviewResult{Reviews: []Review{}}
	sum := 0
	for rows.Next() {
		var v Review
		var jid, title, uid, first, last, image string
		if e = rows.Scan(&v.ID, &jid, &title, &uid, &first, &last, &image, &v.Rating, &v.Comment, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return out, e
		}
		v.Job = map[string]any{"_id": jid, "jobTitle": title}
		v.Reviewer = map[string]any{"_id": uid, "firstName": first, "lastName": last, "profileImage": image}
		v.Reviewee = id
		sum += v.Rating
		out.Reviews = append(out.Reviews, v)
	}
	out.TotalReviews = len(out.Reviews)
	if out.TotalReviews > 0 {
		out.AverageRating = float64(int(float64(sum)/float64(out.TotalReviews)*10+.5)) / 10
	}
	return out, rows.Err()
}
