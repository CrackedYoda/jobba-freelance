package marketplace

import (
	"context"
	"strings"
)

type service struct{ repo Repository }

func NewService(r Repository) Service { return &service{r} }
func (s *service) SearchWorkers(c context.Context, q Search) (WorkerPage, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	return s.repo.SearchWorkers(c, q)
}
func (s *service) Worker(c context.Context, id string) (Worker, error) { return s.repo.Worker(c, id) }
func (s *service) ToggleSaved(c context.Context, client, worker string) (bool, error) {
	if client == worker {
		return false, ErrWorkerNotFound
	}
	return s.repo.ToggleSaved(c, client, worker)
}
func (s *service) Profile(c context.Context, id string) (PublicProfile, error) {
	return s.repo.Profile(c, id)
}
func (s *service) UpsertProfile(c context.Context, id string, in ProfileInput) (PublicProfile, error) {
	if in.Title != nil && len(strings.TrimSpace(*in.Title)) > 100 {
		return PublicProfile{}, ErrInvalid
	}
	if in.Bio != nil && len(strings.TrimSpace(*in.Bio)) > 2000 {
		return PublicProfile{}, ErrInvalid
	}
	if in.HourlyRate != nil && *in.HourlyRate < 0 {
		return PublicProfile{}, ErrInvalid
	}
	return s.repo.UpsertProfile(c, id, in)
}
func (s *service) Recommended(c context.Context, id string, limit int) ([]Job, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.Recommended(c, id, limit)
}
func (s *service) CreateOffer(c context.Context, id string, in OfferInput) (Offer, error) {
	if in.Budget < 0 || in.JobID == "" || in.WorkerID == "" || (in.StartDate != nil && in.EndDate != nil && !in.EndDate.After(*in.StartDate)) {
		return Offer{}, ErrInvalid
	}
	return s.repo.CreateOffer(c, id, in)
}
func (s *service) Offers(c context.Context, id, role string, q OfferQuery) (OfferPage, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		return OfferPage{}, ErrInvalid
	}
	q.Status = strings.ToUpper(q.Status)
	if q.Status != "" && q.Status != "PENDING" && q.Status != "ACCEPTED" && q.Status != "REJECTED" {
		return OfferPage{}, ErrInvalid
	}
	return s.repo.Offers(c, id, role, q)
}
func (s *service) Offer(c context.Context, id, role, offer string) (Offer, error) {
	return s.repo.Offer(c, id, role, offer)
}
func (s *service) AcceptOffer(c context.Context, id, worker string) (Job, error) {
	return s.repo.AcceptOffer(c, id, worker)
}
func (s *service) RejectOffer(c context.Context, id, worker string) error {
	return s.repo.RejectOffer(c, id, worker)
}
func (s *service) CreateReview(c context.Context, user, job string, rating int, comment string) (Review, error) {
	comment = strings.TrimSpace(comment)
	if rating < 1 || rating > 5 || len(comment) > 500 {
		return Review{}, ErrInvalid
	}
	return s.repo.CreateReview(c, user, job, rating, comment)
}
func (s *service) Reviews(c context.Context, id string) (ReviewResult, error) {
	return s.repo.Reviews(c, id)
}
func (s *service) JobContext(c context.Context, user, role, job string) (JobContext, error) {
	return s.repo.JobContext(c, user, role, job)
}
