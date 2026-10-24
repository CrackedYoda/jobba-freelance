package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type repository struct{ pool *pgxpool.Pool }

func NewRepository(p *pgxpool.Pool) Repository { return &repository{p} }
func (r *repository) workerSkills(c context.Context, id string) ([]Skill, error) {
	rows, e := r.pool.Query(c, `SELECT s.id::text,s.name FROM skills s JOIN user_skills ws ON ws.skill_id=s.id WHERE ws.user_id=$1::uuid ORDER BY s.name`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Skill{}
	for rows.Next() {
		var s Skill
		if e = rows.Scan(&s.ID, &s.Name); e != nil {
			return nil, e
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
func scanWorker(row interface{ Scan(...any) error }) (Worker, error) {
	var w Worker
	e := row.Scan(&w.ID, &w.FirstName, &w.LastName, &w.ProfileImage, &w.Country, &w.StateOfResidence, &w.LGA, &w.WorkTypes, &w.CreatedAt)
	w.Skills = []Skill{}
	return w, e
}

const workerCols = `u.id::text,u.first_name,u.last_name,COALESCE(u.profile_image,''),COALESCE(u.country,'Nigeria'),COALESCE(u.state_of_residence,''),COALESCE(u.lga,''),u.work_types,u.created_at`

func (r *repository) SearchWorkers(c context.Context, q Search) (WorkerPage, error) {
	where := []string{`u.account_type='worker'`, `u.is_active=true`, `u.status='ACTIVE'`}
	args := []any{}
	add := func(s string, v any) { args = append(args, v); where = append(where, fmt.Sprintf(s, len(args))) }
	if q.Search != "" {
		args = append(args, q.Search)
		where = append(where, fmt.Sprintf(`(u.first_name ILIKE '%%'||$%d||'%%' OR u.last_name ILIKE '%%'||$%d||'%%')`, len(args), len(args)))
	}
	if q.WorkType != "" {
		add(`$%d=ANY(u.work_types)`, q.WorkType)
	}
	if q.Country != "" {
		add(`u.country=$%d`, q.Country)
	}
	if q.State != "" {
		add(`u.state_of_residence=$%d`, q.State)
	}
	if q.LGA != "" {
		add(`u.lga ILIKE '%%'||$%d||'%%'`, q.LGA)
	}
	if q.SkillID != "" {
		add(`EXISTS(SELECT 1 FROM user_skills ws WHERE ws.user_id=u.id AND ws.skill_id=$%d::uuid)`, q.SkillID)
	}
	if q.SavedBy != "" {
		add(`EXISTS(SELECT 1 FROM saved_workers sw WHERE sw.worker_id=u.id AND sw.client_id=$%d::uuid)`, q.SavedBy)
	}
	base := ` FROM users u WHERE ` + strings.Join(where, " AND ")
	var total int64
	if e := r.pool.QueryRow(c, `SELECT count(*)`+base, args...).Scan(&total); e != nil {
		return WorkerPage{}, e
	}
	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, e := r.pool.Query(c, `SELECT `+workerCols+base+fmt.Sprintf(` ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if e != nil {
		return WorkerPage{}, e
	}
	defer rows.Close()
	p := WorkerPage{TotalCount: total, Data: []Worker{}}
	for rows.Next() {
		w, e := scanWorker(rows)
		if e != nil {
			return p, e
		}
		w.Skills, e = r.workerSkills(c, w.ID)
		if e != nil {
			return p, e
		}
		p.Data = append(p.Data, w)
	}
	p.Count = len(p.Data)
	p.Pages = int((total + int64(q.Limit) - 1) / int64(q.Limit))
	if q.Page < p.Pages {
		v := q.Page + 1
		p.Next = &v
	}
	if q.Page > 1 && p.Pages > 0 {
		v := q.Page - 1
		p.Prev = &v
	}
	return p, rows.Err()
}
func (r *repository) Worker(c context.Context, id string) (Worker, error) {
	w, e := scanWorker(r.pool.QueryRow(c, `SELECT `+workerCols+` FROM users u WHERE u.id=$1::uuid AND u.account_type='worker' AND u.is_active=true`, id))
	if errors.Is(e, pgx.ErrNoRows) {
		return w, ErrWorkerNotFound
	}
	if e == nil {
		w.Skills, e = r.workerSkills(c, id)
	}
	return w, e
}
func (r *repository) ToggleSaved(c context.Context, client, worker string) (bool, error) {
	if _, e := r.Worker(c, worker); e != nil {
		return false, e
	}
	tx, e := r.pool.Begin(c)
	if e != nil {
		return false, e
	}
	defer func() { _ = tx.Rollback(c) }()
	var exists bool
	if e = tx.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM saved_workers WHERE client_id=$1::uuid AND worker_id=$2::uuid)`, client, worker).Scan(&exists); e != nil {
		return false, e
	}
	if exists {
		_, e = tx.Exec(c, `DELETE FROM saved_workers WHERE client_id=$1::uuid AND worker_id=$2::uuid`, client, worker)
	} else {
		_, e = tx.Exec(c, `INSERT INTO saved_workers(client_id,worker_id) VALUES($1::uuid,$2::uuid)`, client, worker)
	}
	if e != nil {
		return false, e
	}
	return !exists, tx.Commit(c)
}
func (r *repository) Profile(c context.Context, id string) (PublicProfile, error) {
	w, e := r.Worker(c, id)
	if e != nil {
		return PublicProfile{}, e
	}
	var title, bio, availability string
	var rate *float64
	var langs, portfolio, education, certs []byte
	e = r.pool.QueryRow(c, `SELECT title,bio,hourly_rate::float8,availability,languages,portfolio,education,certifications FROM worker_profiles WHERE worker_id=$1::uuid`, id).Scan(&title, &bio, &rate, &availability, &langs, &portfolio, &education, &certs)
	if errors.Is(e, pgx.ErrNoRows) {
		title, bio, availability = "", "", "Available for work"
		langs = []byte(`[]`)
		portfolio = []byte(`[]`)
		education = []byte(`[]`)
		certs = []byte(`[]`)
	} else if e != nil {
		return PublicProfile{}, e
	}
	var rating float64
	var reviews, jobs int
	if e = r.pool.QueryRow(c, `SELECT COALESCE(round(avg(rating)::numeric,1),0)::float8,count(*) FROM reviews WHERE reviewee_id=$1::uuid`, id).Scan(&rating, &reviews); e != nil {
		return PublicProfile{}, e
	}
	_ = r.pool.QueryRow(c, `SELECT count(*) FROM jobs WHERE assigned_worker_id=$1::uuid AND status='COMPLETED'`, id).Scan(&jobs)
	decode := func(b []byte) any { var v any; _ = json.Unmarshal(b, &v); return v }
	score := 0
	if rating > 0 {
		score = int(rating/5*100 + .5)
	}
	return PublicProfile{Worker: map[string]any{"id": w.ID, "firstName": w.FirstName, "lastName": w.LastName, "avatarUrl": w.ProfileImage, "title": title, "location": map[string]string{"city": w.LGA, "state": w.StateOfResidence, "country": w.Country}, "bio": bio, "hourlyRate": rate, "joinedDate": w.CreatedAt}, Stats: map[string]any{"jobSuccessScore": score, "rating": rating, "jobsCompleted": jobs, "totalReviews": reviews}, Skills: w.Skills, WorkHistory: []any{}, Portfolio: decode(portfolio), Languages: decode(langs), Availability: availability, Education: decode(education), Certifications: decode(certs)}, nil
}
func (r *repository) UpsertProfile(c context.Context, id string, in ProfileInput) (PublicProfile, error) {
	if _, e := r.Worker(c, id); e != nil {
		return PublicProfile{}, e
	}
	empty := []byte(`[]`)
	langs, port, edu, cert := in.Languages, in.Portfolio, in.Education, in.Certifications
	if len(langs) == 0 {
		langs = empty
	}
	if len(port) == 0 {
		port = empty
	}
	if len(edu) == 0 {
		edu = empty
	}
	if len(cert) == 0 {
		cert = empty
	}
	_, e := r.pool.Exec(c, `INSERT INTO worker_profiles(worker_id,title,bio,hourly_rate,availability,languages,portfolio,education,certifications) VALUES($1::uuid,COALESCE($2,''),COALESCE($3,''),$4,COALESCE($5,'Available for work'),$6,$7,$8,$9) ON CONFLICT(worker_id) DO UPDATE SET title=COALESCE($2,worker_profiles.title),bio=COALESCE($3,worker_profiles.bio),hourly_rate=COALESCE($4,worker_profiles.hourly_rate),availability=COALESCE($5,worker_profiles.availability),languages=CASE WHEN $6::jsonb='[]' THEN worker_profiles.languages ELSE $6 END,portfolio=CASE WHEN $7::jsonb='[]' THEN worker_profiles.portfolio ELSE $7 END,education=CASE WHEN $8::jsonb='[]' THEN worker_profiles.education ELSE $8 END,certifications=CASE WHEN $9::jsonb='[]' THEN worker_profiles.certifications ELSE $9 END,updated_at=now()`, id, in.Title, in.Bio, in.HourlyRate, in.Availability, langs, port, edu, cert)
	if e != nil {
		return PublicProfile{}, e
	}
	return r.Profile(c, id)
}
func (r *repository) Recommended(c context.Context, id string, limit int) ([]Job, error) {
	q := `SELECT j.id::text,j.job_title,j.category,j.description,j.min_budget::float8,j.max_budget::float8,j.work_type::text,j.location,j.status::text,j.client_id::text,j.created_at FROM jobs j JOIN users u ON u.id=$1::uuid WHERE j.status='OPEN' AND NOT j.is_deleted AND j.privacy='public' AND j.assigned_worker_id IS NULL AND EXISTS(SELECT 1 FROM job_skills js JOIN user_skills ws ON ws.skill_id=js.skill_id WHERE js.job_id=j.id AND ws.user_id=u.id) AND (j.work_type='remote' OR (j.work_type='onsite' AND (COALESCE(u.state_of_residence,'')<>'' AND j.location ILIKE '%%'||u.state_of_residence||'%%' OR COALESCE(u.lga,'')<>'' AND j.location ILIKE '%%'||u.lga||'%%'))) AND NOT EXISTS(SELECT 1 FROM applications a WHERE a.job_id=j.id AND a.worker_id=u.id) ORDER BY j.created_at DESC LIMIT $2`
	rows, e := r.pool.Query(c, q, id, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Job{}
	for rows.Next() {
		var j Job
		if e = rows.Scan(&j.ID, &j.JobTitle, &j.Category, &j.Description, &j.MinBudget, &j.MaxBudget, &j.WorkType, &j.Location, &j.Status, &j.ClientID, &j.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
