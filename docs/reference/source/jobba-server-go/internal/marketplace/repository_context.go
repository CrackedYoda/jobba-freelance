package marketplace

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func (r *repository) Offers(c context.Context, user, role string, q OfferQuery) (OfferPage, error) {
	column := "client_id"
	if role == "worker" {
		column = "worker_id"
	}
	where := fmt.Sprintf(`%s=$1::uuid`, column)
	args := []any{user}
	if q.Status != "" {
		args = append(args, q.Status)
		where += fmt.Sprintf(` AND status=$%d`, len(args))
	}
	var total int64
	if e := r.pool.QueryRow(c, `SELECT count(*) FROM offers WHERE `+where, args...).Scan(&total); e != nil {
		return OfferPage{}, e
	}
	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, e := r.pool.Query(c, `SELECT `+offerCols+` FROM offers WHERE `+where+fmt.Sprintf(` ORDER BY created_at DESC,id DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if e != nil {
		return OfferPage{}, e
	}
	defer rows.Close()
	p := OfferPage{Docs: []Offer{}, TotalDocs: total, Limit: q.Limit, Page: q.Page}
	p.TotalPages = int((total + int64(q.Limit) - 1) / int64(q.Limit))
	p.HasPrevPage = q.Page > 1 && p.TotalPages > 0
	p.HasNextPage = q.Page < p.TotalPages
	if p.HasPrevPage {
		v := q.Page - 1
		p.PrevPage = &v
	}
	if p.HasNextPage {
		v := q.Page + 1
		p.NextPage = &v
	}
	for rows.Next() {
		o, e := scanOffer(rows)
		if e != nil {
			return p, e
		}
		p.Docs = append(p.Docs, o)
	}
	return p, rows.Err()
}
func (r *repository) Offer(c context.Context, user, role, id string) (Offer, error) {
	column := "client_id"
	if role == "worker" {
		column = "worker_id"
	}
	o, e := scanOffer(r.pool.QueryRow(c, `SELECT `+offerCols+` FROM offers WHERE id=$1::uuid AND `+column+`=$2::uuid`, id, user))
	if errors.Is(e, pgx.ErrNoRows) {
		return Offer{}, ErrOfferNotFound
	}
	return o, e
}
func (r *repository) JobContext(c context.Context, user, role, jobID string) (JobContext, error) {
	var client string
	var worker *string
	var status string
	e := r.pool.QueryRow(c, `SELECT client_id::text,assigned_worker_id::text,status::text FROM jobs WHERE id=$1::uuid AND NOT is_deleted`, jobID).Scan(&client, &worker, &status)
	if errors.Is(e, pgx.ErrNoRows) {
		return JobContext{}, ErrJobNotFound
	}
	if e != nil {
		return JobContext{}, e
	}
	participant := client == user || (worker != nil && *worker == user)
	if !participant {
		return JobContext{}, ErrReviewForbidden
	}
	ctx := JobContext{JobID: jobID, Status: status, Role: role, ReviewEndpoint: "/api/v1/jobs/" + role + "/" + jobID + "/review", DisputeEndpoint: "/api/v1/disputes"}
	_ = r.pool.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM reviews WHERE job_id=$1::uuid AND reviewer_id=$2::uuid)`, jobID, user).Scan(&ctx.HasReviewed)
	ctx.CanReview = status == "COMPLETED" && !ctx.HasReviewed
	_ = r.pool.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM disputes WHERE job_id=$1::uuid AND status IN ('OPEN','UNDER_REVIEW'))`, jobID).Scan(&ctx.HasOpenDispute)
	ctx.CanDispute = status == "ACTIVE" && !ctx.HasOpenDispute
	other := client
	if client == user && worker != nil {
		other = *worker
	}
	if other != user {
		var first, last, image string
		if e := r.pool.QueryRow(c, `SELECT first_name,last_name,COALESCE(profile_image,'') FROM users WHERE id=$1::uuid`, other).Scan(&first, &last, &image); e == nil {
			ctx.Counterparty = map[string]any{"_id": other, "firstName": first, "lastName": last, "profileImage": image}
		}
	}
	return ctx, nil
}
