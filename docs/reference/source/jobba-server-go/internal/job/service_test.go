package job

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	current      Job
	created      CreateInput
	filter       ListFilter
	err          error
	status       Status
	saved        bool
	workerExists bool
	canView      bool
}

func (f *fakeRepository) Create(_ context.Context, _ string, in CreateInput) (Job, error) {
	f.created = in
	return f.current, f.err
}
func (f *fakeRepository) List(_ context.Context, in ListFilter) (Page, error) {
	f.filter = in
	return Page{}, f.err
}
func (f *fakeRepository) GetOwned(context.Context, string, string) (Job, error) {
	return f.current, f.err
}
func (f *fakeRepository) GetVisible(context.Context, string) (Job, error) { return f.current, f.err }
func (f *fakeRepository) WorkerCanView(context.Context, string, string) (bool, error) {
	return f.canView, f.err
}
func (f *fakeRepository) Update(_ context.Context, _ Job, _ UpdateInput) (Job, error) {
	return f.current, f.err
}
func (f *fakeRepository) SoftDelete(context.Context, string) error { return f.err }
func (f *fakeRepository) SetStatus(_ context.Context, _ string, s Status) (Job, error) {
	f.status = s
	return f.current, f.err
}
func (f *fakeRepository) Assign(context.Context, string, string) (Job, error) {
	return f.current, f.err
}
func (f *fakeRepository) WorkerExists(context.Context, string) (bool, error) {
	return f.workerExists, f.err
}
func (f *fakeRepository) IncrementView(context.Context, string) (Job, error) {
	f.current.ViewCount++
	return f.current, f.err
}
func (f *fakeRepository) Statistics(context.Context, string) (Statistics, error) {
	return Statistics{}, f.err
}
func (f *fakeRepository) Save(context.Context, string, string) error   { f.saved = true; return f.err }
func (f *fakeRepository) Unsave(context.Context, string, string) error { f.saved = false; return f.err }
func (f *fakeRepository) ListSaved(context.Context, string, Pagination) (Page, error) {
	return Page{}, f.err
}

func validCreate(now time.Time) CreateInput {
	return CreateInput{JobTitle: "Build a garden wall", Category: "Masonry", Description: "Build a durable garden wall with supplied blocks and a clean finish.", MinBudget: 100, MaxBudget: 200, PaymentMethod: PaymentEscrow, StartDate: now.Add(time.Hour), EndDate: now.Add(48 * time.Hour), WorkType: WorkOnsite, Location: "Lagos", ApplicantType: ApplicantAll, Privacy: PrivacyPublic}
}
func TestCreateValidatesBusinessRules(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*CreateInput)
	}{{"past start", func(i *CreateInput) { i.StartDate = now.Add(-time.Hour) }}, {"inverted budget", func(i *CreateInput) { i.MaxBudget = i.MinBudget - 1 }}, {"onsite location", func(i *CreateInput) { i.Location = "" }}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{}
			svc := NewService(repo).(*service)
			svc.now = func() time.Time { return now }
			in := validCreate(now)
			tt.mutate(&in)
			if _, err := svc.Create(context.Background(), "client", in); !errors.Is(err, ErrInvalid) {
				t.Fatalf("expected invalid job, got %v", err)
			}
		})
	}
}
func TestSearchForcesPublicOpenUnassigned(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	_, err := svc.Search(context.Background(), ListFilter{Pagination: Pagination{Page: 1, Limit: 10}, SortBy: "createdAt", Order: "desc"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.filter.PublicOnly || !repo.filter.UnassignedOnly || repo.filter.Status == nil || *repo.filter.Status != StatusOpen {
		t.Fatalf("unsafe search filter: %#v", repo.filter)
	}
}
func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name     string
		current  Status
		complete bool
		wantErr  error
		want     Status
	}{{"complete active", StatusActive, true, nil, StatusCompleted}, {"reject complete open", StatusOpen, true, ErrNotActive, ""}, {"close open", StatusOpen, false, nil, StatusClosed}, {"reject close completed", StatusCompleted, false, ErrAlreadyCompleted, ""}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{current: Job{Status: tt.current}}
			svc := NewService(repo)
			var err error
			if tt.complete {
				_, err = svc.Complete(context.Background(), "job", "client")
			} else {
				_, err = svc.Close(context.Background(), "job", "client")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v want %v", err, tt.wantErr)
			}
			if tt.want != "" && repo.status != tt.want {
				t.Fatalf("got transition %s", repo.status)
			}
		})
	}
}

func TestAssignRejectsNonWorker(t *testing.T) {
	repo := &fakeRepository{current: Job{Status: StatusOpen}, workerExists: false}
	_, err := NewService(repo).Assign(context.Background(), "job", "client", "not-worker")
	if !errors.Is(err, ErrInvalidWorker) {
		t.Fatalf("expected invalid worker, got %v", err)
	}
}
func TestPrivateJobDetailsRequireInvitation(t *testing.T) {
	repo := &fakeRepository{current: Job{Privacy: PrivacyPrivate, ApplicantType: ApplicantInvite}}
	_, err := NewService(repo).GetWorker(context.Background(), "job", "worker")
	if !errors.Is(err, ErrPrivate) {
		t.Fatalf("expected private error, got %v", err)
	}
	repo.canView = true
	if _, err = NewService(repo).GetWorker(context.Background(), "job", "worker"); err != nil {
		t.Fatalf("invited worker rejected: %v", err)
	}
}
