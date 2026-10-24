package job

import (
	"context"
	"errors"
	"strings"
	"time"
)

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) Create(ctx context.Context, clientID string, in CreateInput) (Job, error) {
	normalizeCreate(&in)
	if err := validateCreate(in, s.now()); err != nil {
		return Job{}, err
	}
	job, err := s.repository.Create(ctx, clientID, in)
	return job, wrapRepo(err)
}
func (s *service) ListClient(ctx context.Context, id string, f ListFilter) (Page, error) {
	f.ClientID = id
	return s.list(ctx, f)
}
func (s *service) ListWorker(ctx context.Context, id string, f ListFilter) (Page, error) {
	f.WorkerID = id
	return s.list(ctx, f)
}
func (s *service) Search(ctx context.Context, f ListFilter) (Page, error) {
	f.PublicOnly = true
	f.UnassignedOnly = true
	open := StatusOpen
	f.Status = &open
	return s.list(ctx, f)
}
func (s *service) ByCategory(ctx context.Context, category string, p Pagination) (Page, error) {
	return s.Search(ctx, ListFilter{Pagination: p, Category: strings.TrimSpace(category), SortBy: "createdAt", Order: "desc"})
}
func (s *service) list(ctx context.Context, f ListFilter) (Page, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 50
	}
	if f.SortBy == "" {
		f.SortBy = "createdAt"
	}
	if f.Order == "" {
		f.Order = "desc"
	}
	page, err := s.repository.List(ctx, f)
	return page, wrapRepo(err)
}
func (s *service) GetClient(ctx context.Context, id, client string) (Job, error) {
	j, e := s.repository.GetOwned(ctx, id, client)
	return j, wrapRepo(e)
}
func (s *service) GetWorker(ctx context.Context, id, workerID string) (Job, error) {
	j, e := s.repository.GetVisible(ctx, id)
	if e != nil {
		return Job{}, wrapRepo(e)
	}
	if j.Privacy == PrivacyPrivate {
		if j.ApplicantType != ApplicantInvite {
			return Job{}, ErrPrivate
		}
		allowed, err := s.repository.WorkerCanView(ctx, id, workerID)
		if err != nil {
			return Job{}, wrapRepo(err)
		}
		if !allowed {
			return Job{}, ErrPrivate
		}
	}
	j, e = s.repository.IncrementView(ctx, id)
	return j, wrapRepo(e)
}
func (s *service) Update(ctx context.Context, id, client string, in UpdateInput) (Job, error) {
	current, e := s.repository.GetOwned(ctx, id, client)
	if e != nil {
		return Job{}, wrapRepo(e)
	}
	if current.Status == StatusActive || current.Status == StatusCompleted {
		return Job{}, ErrImmutable
	}
	if e = validateUpdate(current, in); e != nil {
		return Job{}, e
	}
	j, e := s.repository.Update(ctx, current, in)
	return j, wrapRepo(e)
}
func (s *service) Delete(ctx context.Context, id, client string) error {
	j, e := s.repository.GetOwned(ctx, id, client)
	if e != nil {
		return wrapRepo(e)
	}
	if j.Status == StatusActive {
		return ErrDeleteActive
	}
	return wrapRepo(s.repository.SoftDelete(ctx, id))
}
func (s *service) Close(ctx context.Context, id, client string) (Job, error) {
	j, e := s.repository.GetOwned(ctx, id, client)
	if e != nil {
		return Job{}, wrapRepo(e)
	}
	if j.Status == StatusCompleted {
		return Job{}, ErrAlreadyCompleted
	}
	j, e = s.repository.SetStatus(ctx, id, StatusClosed)
	return j, wrapRepo(e)
}
func (s *service) Complete(ctx context.Context, id, client string) (Job, error) {
	j, e := s.repository.GetOwned(ctx, id, client)
	if e != nil {
		return Job{}, wrapRepo(e)
	}
	if j.Status != StatusActive {
		return Job{}, ErrNotActive
	}
	j, e = s.repository.SetStatus(ctx, id, StatusCompleted)
	return j, wrapRepo(e)
}
func (s *service) Assign(ctx context.Context, id, client, worker string) (Job, error) {
	j, e := s.repository.GetOwned(ctx, id, client)
	if e != nil {
		return Job{}, wrapRepo(e)
	}
	if j.Status != StatusOpen {
		return Job{}, ErrNotOpen
	}
	exists, e := s.repository.WorkerExists(ctx, worker)
	if e != nil {
		return Job{}, wrapRepo(e)
	}
	if !exists {
		return Job{}, ErrInvalidWorker
	}
	j, e = s.repository.Assign(ctx, id, worker)
	return j, wrapRepo(e)
}
func (s *service) Statistics(ctx context.Context, id string) (Statistics, error) {
	v, e := s.repository.Statistics(ctx, id)
	return v, wrapRepo(e)
}
func (s *service) Save(ctx context.Context, user, id string) error {
	if _, e := s.repository.GetVisible(ctx, id); e != nil {
		return wrapRepo(e)
	}
	return wrapRepo(s.repository.Save(ctx, user, id))
}
func (s *service) Unsave(ctx context.Context, user, id string) error {
	if _, e := s.repository.GetVisible(ctx, id); e != nil {
		return wrapRepo(e)
	}
	return wrapRepo(s.repository.Unsave(ctx, user, id))
}
func (s *service) Saved(ctx context.Context, user string, p Pagination) (Page, error) {
	v, e := s.repository.ListSaved(ctx, user, p)
	return v, wrapRepo(e)
}

func normalizeCreate(in *CreateInput) {
	in.JobTitle = strings.TrimSpace(in.JobTitle)
	in.Category = strings.TrimSpace(in.Category)
	in.Description = strings.TrimSpace(in.Description)
	in.Location = strings.TrimSpace(in.Location)
	if in.SkillIDs == nil {
		in.SkillIDs = []string{}
	}
	if in.Attachments == nil {
		in.Attachments = []string{}
	}
}
func validateCreate(in CreateInput, now time.Time) error {
	if len(in.JobTitle) < 5 || len(in.JobTitle) > 100 || in.Category == "" || len(in.Description) < 50 || len(in.Description) > 2000 {
		return ErrInvalid
	}
	if in.MinBudget < 0 || in.MaxBudget < in.MinBudget {
		return ErrInvalid
	}
	if in.PaymentMethod != PaymentEscrow && in.PaymentMethod != PaymentDirect {
		return ErrInvalid
	}
	if !in.StartDate.After(now.Add(-time.Minute)) || !in.EndDate.After(in.StartDate) {
		return ErrInvalid
	}
	if in.WorkType != WorkOnsite && in.WorkType != WorkRemote {
		return ErrInvalid
	}
	if in.WorkType == WorkOnsite && in.Location == "" {
		return ErrInvalid
	}
	if in.ApplicantType != ApplicantAll && in.ApplicantType != ApplicantTopRated && in.ApplicantType != ApplicantInvite {
		return ErrInvalid
	}
	if in.Privacy != PrivacyPublic && in.Privacy != PrivacyPrivate {
		return ErrInvalid
	}
	for _, id := range in.SkillIDs {
		if !validID(id) {
			return ErrInvalid
		}
	}
	return nil
}
func validateUpdate(j Job, in UpdateInput) error {
	start, end, min, max := j.StartDate, j.EndDate, j.MinBudget, j.MaxBudget
	if in.StartDate != nil {
		start = *in.StartDate
	}
	if in.EndDate != nil {
		end = *in.EndDate
	}
	if !end.After(start) {
		return ErrInvalid
	}
	if in.MinBudget != nil {
		min = *in.MinBudget
	}
	if in.MaxBudget != nil {
		max = *in.MaxBudget
	}
	if min < 0 || max < min {
		return ErrInvalid
	}
	if in.JobTitle != nil && (len(strings.TrimSpace(*in.JobTitle)) < 5 || len(strings.TrimSpace(*in.JobTitle)) > 100) {
		return ErrInvalid
	}
	if in.Description != nil && (len(strings.TrimSpace(*in.Description)) < 50 || len(strings.TrimSpace(*in.Description)) > 2000) {
		return ErrInvalid
	}
	if in.PaymentMethod != nil && *in.PaymentMethod != PaymentEscrow && *in.PaymentMethod != PaymentDirect {
		return ErrInvalid
	}
	if in.WorkType != nil && *in.WorkType != WorkOnsite && *in.WorkType != WorkRemote {
		return ErrInvalid
	}
	if in.ApplicantType != nil && *in.ApplicantType != ApplicantAll && *in.ApplicantType != ApplicantTopRated && *in.ApplicantType != ApplicantInvite {
		return ErrInvalid
	}
	if in.Privacy != nil && *in.Privacy != PrivacyPublic && *in.Privacy != PrivacyPrivate {
		return ErrInvalid
	}
	if in.Status != nil && !validStatus(*in.Status) {
		return ErrInvalid
	}
	if in.SkillIDs != nil {
		for _, id := range *in.SkillIDs {
			if !validID(id) {
				return ErrInvalid
			}
		}
	}
	work, location := j.WorkType, j.Location
	if in.WorkType != nil {
		work = *in.WorkType
	}
	if in.Location != nil {
		location = strings.TrimSpace(*in.Location)
	}
	if work == WorkOnsite && location == "" {
		return ErrInvalid
	}
	return nil
}
func wrapRepo(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return ErrNotFound
	}
	return ErrRepository.WithError(err)
}
