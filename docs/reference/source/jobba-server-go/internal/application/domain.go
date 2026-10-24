package application

import "time"

type Status string

const (
	StatusPending     Status = "PENDING"
	StatusShortlisted Status = "SHORTLISTED"
	StatusRejected    Status = "REJECTED"
	StatusAccepted    Status = "ACCEPTED"
)

type UserSummary struct {
	ID           string `json:"_id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	ProfileImage string `json:"profileImage,omitempty"`
}

type JobSummary struct {
	ID        string      `json:"_id"`
	JobTitle  string      `json:"jobTitle"`
	MinBudget float64     `json:"minBudget"`
	MaxBudget float64     `json:"maxBudget"`
	WorkType  string      `json:"workType"`
	Status    string      `json:"status"`
	Client    UserSummary `json:"client"`
}

type Application struct {
	ID           string    `json:"_id"`
	JobID        string    `json:"-"`
	WorkerID     string    `json:"-"`
	Job          any       `json:"job"`
	Worker       any       `json:"worker"`
	CoverLetter  string    `json:"coverLetter"`
	BidAmount    float64   `json:"bidAmount"`
	WorkDuration string    `json:"workDuration"`
	Status       Status    `json:"status"`
	Attachments  []string  `json:"attachments"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Page struct {
	Docs          []Application `json:"docs"`
	TotalDocs     int64         `json:"totalDocs"`
	Limit         int           `json:"limit"`
	TotalPages    int           `json:"totalPages"`
	Page          int           `json:"page"`
	PagingCounter int           `json:"pagingCounter"`
	HasPrevPage   bool          `json:"hasPrevPage"`
	HasNextPage   bool          `json:"hasNextPage"`
	PrevPage      *int          `json:"prevPage"`
	NextPage      *int          `json:"nextPage"`
}

type JobState struct {
	ID, ClientID  string
	Status        string
	IsDeleted     bool
	Privacy       string
	ApplicantType string
}
