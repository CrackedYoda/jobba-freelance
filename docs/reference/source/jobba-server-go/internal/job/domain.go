package job

import "time"

type Status string
type PaymentMethod string
type WorkType string
type ApplicantType string
type Privacy string

const (
	StatusDraft       Status        = "DRAFT"
	StatusOpen        Status        = "OPEN"
	StatusDisputed    Status        = "DISPUTED"
	StatusActive      Status        = "ACTIVE"
	StatusCompleted   Status        = "COMPLETED"
	StatusCancelled   Status        = "CANCELLED"
	StatusClosed      Status        = "CLOSED"
	PaymentEscrow     PaymentMethod = "escrow"
	PaymentDirect     PaymentMethod = "direct"
	WorkOnsite        WorkType      = "onsite"
	WorkRemote        WorkType      = "remote"
	ApplicantAll      ApplicantType = "all"
	ApplicantTopRated ApplicantType = "top-rated"
	ApplicantInvite   ApplicantType = "invite"
	PrivacyPublic     Privacy       = "public"
	PrivacyPrivate    Privacy       = "private"
)

type Skill struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

type Job struct {
	ID               string        `json:"_id"`
	UniqueID         string        `json:"uniqueId,omitempty"`
	ClientID         string        `json:"client"`
	JobTitle         string        `json:"jobTitle"`
	Category         string        `json:"category"`
	Description      string        `json:"description"`
	Skills           []Skill       `json:"skills"`
	Attachments      []string      `json:"attachments"`
	MinBudget        float64       `json:"minBudget"`
	MaxBudget        float64       `json:"maxBudget"`
	PaymentMethod    PaymentMethod `json:"paymentMethod"`
	StartDate        time.Time     `json:"startDate"`
	EndDate          time.Time     `json:"endDate"`
	WorkType         WorkType      `json:"workType"`
	Location         string        `json:"location"`
	ApplicantType    ApplicantType `json:"applicantType"`
	Privacy          Privacy       `json:"privacy"`
	Status           Status        `json:"status"`
	AssignedWorkerID *string       `json:"assignedWorker,omitempty"`
	Budget           *float64      `json:"budget,omitempty"`
	ViewCount        int64         `json:"viewCount"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
}

type Page struct {
	Count      int   `json:"count"`
	TotalCount int64 `json:"totalCount"`
	Pages      int   `json:"pages"`
	Next       *int  `json:"next,omitempty"`
	Prev       *int  `json:"prev,omitempty"`
	Data       []Job `json:"data"`
}
type LegacyPage struct {
	Docs          []Job `json:"docs"`
	TotalDocs     int64 `json:"totalDocs"`
	Limit         int   `json:"limit"`
	TotalPages    int   `json:"totalPages"`
	Page          int   `json:"page"`
	PagingCounter int   `json:"pagingCounter"`
	HasPrevPage   bool  `json:"hasPrevPage"`
	HasNextPage   bool  `json:"hasNextPage"`
	PrevPage      *int  `json:"prevPage"`
	NextPage      *int  `json:"nextPage"`
}

func (p Page) Legacy(page, limit int) LegacyPage {
	return LegacyPage{Docs: p.Data, TotalDocs: p.TotalCount, Limit: limit, TotalPages: p.Pages, Page: page, PagingCounter: (page-1)*limit + 1, HasPrevPage: p.Prev != nil, HasNextPage: p.Next != nil, PrevPage: p.Prev, NextPage: p.Next}
}

type StatusCount struct {
	ID    Status `json:"_id"`
	Count int64  `json:"count"`
}
type Statistics struct {
	TotalJobs       int64         `json:"totalJobs"`
	StatusBreakdown []StatusCount `json:"statusBreakdown"`
	TotalSpent      float64       `json:"totalSpent"`
}
