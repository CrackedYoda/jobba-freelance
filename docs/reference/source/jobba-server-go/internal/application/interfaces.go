package application

import "context"

type Repository interface {
	GetJob(context.Context, string) (JobState, error)
	HasInvitation(context.Context, string, string) (bool, error)
	Create(context.Context, string, string, CreateInput) (Application, error)
	ListWorker(context.Context, string, ListInput) (Page, error)
	ListJob(context.Context, string, ListInput) (Page, error)
	Accept(context.Context, string, string) (Application, error)
}

type Service interface {
	Apply(context.Context, string, string, CreateInput) (Application, error)
	WorkerApplications(context.Context, string, ListInput) (Page, error)
	JobApplications(context.Context, string, string, ListInput) (Page, error)
	Accept(context.Context, string, string) (Application, error)
}

type CreateInput struct {
	CoverLetter  string   `json:"coverLetter"`
	BidAmount    float64  `json:"bidAmount"`
	WorkDuration string   `json:"workDuration"`
	Attachments  []string `json:"attachments"`
}

type ListInput struct {
	Page, Limit int
	Status      *Status
}
