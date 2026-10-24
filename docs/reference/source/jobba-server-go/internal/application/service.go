package application

import (
	"context"
	"errors"
	"strings"

	"github.com/jobba/backend/internal/shared"
)

type service struct{ repository Repository }

func NewService(repository Repository) Service { return &service{repository: repository} }

func (s *service) Apply(ctx context.Context, workerID, jobID string, in CreateInput) (Application, error) {
	in.CoverLetter, in.WorkDuration = strings.TrimSpace(in.CoverLetter), strings.TrimSpace(in.WorkDuration)
	if in.Attachments == nil {
		in.Attachments = []string{}
	}
	if len(in.CoverLetter) < 10 || in.BidAmount < 1 || in.WorkDuration == "" {
		return Application{}, ErrInvalid
	}
	job, err := s.repository.GetJob(ctx, jobID)
	if err != nil {
		return Application{}, wrap(err)
	}
	if job.IsDeleted || job.Status != "OPEN" {
		return Application{}, ErrJobInactive
	}
	if job.ClientID == workerID {
		return Application{}, ErrOwnJob
	}
	if job.Privacy == "private" {
		if job.ApplicantType != "invite" {
			return Application{}, ErrInvitationRequired
		}
		invited, err := s.repository.HasInvitation(ctx, jobID, workerID)
		if err != nil {
			return Application{}, wrap(err)
		}
		if !invited {
			return Application{}, ErrInvitationRequired
		}
	}
	a, err := s.repository.Create(ctx, jobID, workerID, in)
	return a, wrap(err)
}

func (s *service) WorkerApplications(ctx context.Context, workerID string, in ListInput) (Page, error) {
	p, err := s.repository.ListWorker(ctx, workerID, normalizeList(in))
	return p, wrap(err)
}

func (s *service) JobApplications(ctx context.Context, clientID, jobID string, in ListInput) (Page, error) {
	job, err := s.repository.GetJob(ctx, jobID)
	if err != nil {
		return Page{}, wrap(err)
	}
	if job.ClientID != clientID {
		return Page{}, ErrJobUnauthorized
	}
	p, err := s.repository.ListJob(ctx, jobID, normalizeList(in))
	return p, wrap(err)
}

func (s *service) Accept(ctx context.Context, clientID, applicationID string) (Application, error) {
	a, err := s.repository.Accept(ctx, applicationID, clientID)
	return a, wrap(err)
}

func normalizeList(in ListInput) ListInput {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.Limit < 1 {
		in.Limit = 10
	}
	return in
}

func wrap(err error) error {
	if err == nil {
		return nil
	}
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrOperation.WithError(err)
}
