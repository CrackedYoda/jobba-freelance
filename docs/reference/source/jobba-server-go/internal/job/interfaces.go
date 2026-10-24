package job

import "context"

type Repository interface {
	Create(context.Context, string, CreateInput) (Job, error)
	List(context.Context, ListFilter) (Page, error)
	GetOwned(context.Context, string, string) (Job, error)
	GetVisible(context.Context, string) (Job, error)
	WorkerCanView(context.Context, string, string) (bool, error)
	Update(context.Context, Job, UpdateInput) (Job, error)
	SoftDelete(context.Context, string) error
	SetStatus(context.Context, string, Status) (Job, error)
	Assign(context.Context, string, string) (Job, error)
	WorkerExists(context.Context, string) (bool, error)
	IncrementView(context.Context, string) (Job, error)
	Statistics(context.Context, string) (Statistics, error)
	Save(context.Context, string, string) error
	Unsave(context.Context, string, string) error
	ListSaved(context.Context, string, Pagination) (Page, error)
}

type Service interface {
	Create(context.Context, string, CreateInput) (Job, error)
	ListClient(context.Context, string, ListFilter) (Page, error)
	ListWorker(context.Context, string, ListFilter) (Page, error)
	Search(context.Context, ListFilter) (Page, error)
	ByCategory(context.Context, string, Pagination) (Page, error)
	GetClient(context.Context, string, string) (Job, error)
	GetWorker(context.Context, string, string) (Job, error)
	Update(context.Context, string, string, UpdateInput) (Job, error)
	Delete(context.Context, string, string) error
	Close(context.Context, string, string) (Job, error)
	Complete(context.Context, string, string) (Job, error)
	Assign(context.Context, string, string, string) (Job, error)
	Statistics(context.Context, string) (Statistics, error)
	Save(context.Context, string, string) error
	Unsave(context.Context, string, string) error
	Saved(context.Context, string, Pagination) (Page, error)
}

type Pagination struct{ Page, Limit int }
type ListFilter struct {
	Pagination
	ClientID, WorkerID, Category, Search, SortBy, Order string
	Status                                              *Status
	WorkType                                            *WorkType
	MinBudget, MaxBudget                                *float64
	SkillIDs                                            []string
	PublicOnly, UnassignedOnly                          bool
}
