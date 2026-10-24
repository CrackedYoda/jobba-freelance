package application

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	job         JobState
	application Application
	page        Page
	err         error
	accepted    bool
	invited     bool
}

func (f *fakeRepository) GetJob(context.Context, string) (JobState, error) { return f.job, f.err }
func (f *fakeRepository) HasInvitation(context.Context, string, string) (bool, error) {
	return f.invited, f.err
}
func (f *fakeRepository) Create(context.Context, string, string, CreateInput) (Application, error) {
	return f.application, f.err
}
func (f *fakeRepository) ListWorker(context.Context, string, ListInput) (Page, error) {
	return f.page, f.err
}
func (f *fakeRepository) ListJob(context.Context, string, ListInput) (Page, error) {
	return f.page, f.err
}
func (f *fakeRepository) Accept(context.Context, string, string) (Application, error) {
	f.accepted = true
	return f.application, f.err
}

func validInput() CreateInput {
	return CreateInput{CoverLetter: "I have the experience required for this work.", BidAmount: 100, WorkDuration: "3 days"}
}
func TestApplyRules(t *testing.T) {
	tests := []struct {
		name   string
		job    JobState
		worker string
		want   error
	}{{"open job", JobState{ClientID: "client", Status: "OPEN"}, "worker", nil}, {"closed job", JobState{ClientID: "client", Status: "CLOSED"}, "worker", ErrJobInactive}, {"own job", JobState{ClientID: "worker", Status: "OPEN"}, "worker", ErrOwnJob}, {"deleted job", JobState{ClientID: "client", Status: "OPEN", IsDeleted: true}, "worker", ErrJobInactive}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&fakeRepository{job: tt.job})
			_, err := svc.Apply(context.Background(), tt.worker, "job", validInput())
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v want %v", err, tt.want)
			}
		})
	}
}
func TestApplyValidatesPayload(t *testing.T) {
	svc := NewService(&fakeRepository{job: JobState{Status: "OPEN"}})
	for _, in := range []CreateInput{{CoverLetter: "short", BidAmount: 1, WorkDuration: "day"}, {CoverLetter: "long enough cover letter", BidAmount: 0, WorkDuration: "day"}, {CoverLetter: "long enough cover letter", BidAmount: 1}} {
		if _, err := svc.Apply(context.Background(), "worker", "job", in); !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected invalid, got %v", err)
		}
	}
}
func TestJobApplicationsHideOwnership(t *testing.T) {
	svc := NewService(&fakeRepository{job: JobState{ClientID: "another"}})
	_, err := svc.JobApplications(context.Background(), "client", "job", ListInput{})
	if !errors.Is(err, ErrJobUnauthorized) {
		t.Fatalf("expected ownership-safe not found, got %v", err)
	}
}
func TestPrivateJobRequiresInvitation(t *testing.T) {
	repo := &fakeRepository{job: JobState{ClientID: "client", Status: "OPEN", Privacy: "private", ApplicantType: "invite"}}
	_, err := NewService(repo).Apply(context.Background(), "worker", "job", validInput())
	if !errors.Is(err, ErrInvitationRequired) {
		t.Fatalf("expected private job rejection, got %v", err)
	}
	repo.invited = true
	if _, err = NewService(repo).Apply(context.Background(), "worker", "job", validInput()); err != nil {
		t.Fatalf("invited worker rejected: %v", err)
	}
}
func TestPageMatchesMongooseShape(t *testing.T) {
	p := newPage(21, ListInput{Page: 2, Limit: 10})
	if p.TotalPages != 3 || p.PagingCounter != 11 || !p.HasPrevPage || !p.HasNextPage || *p.PrevPage != 1 || *p.NextPage != 3 {
		t.Fatalf("unexpected page: %#v", p)
	}
}
