package marketplace

import (
	"encoding/json"
	"time"
)

type Skill struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}
type Worker struct {
	ID               string    `json:"_id"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	ProfileImage     string    `json:"profileImage"`
	Country          string    `json:"country"`
	StateOfResidence string    `json:"stateOfResidence"`
	LGA              string    `json:"lga"`
	WorkTypes        []string  `json:"workType"`
	Skills           []Skill   `json:"skills"`
	CreatedAt        time.Time `json:"createdAt"`
}
type WorkerPage struct {
	Count      int      `json:"count"`
	TotalCount int64    `json:"totalCount"`
	Pages      int      `json:"pages"`
	Next       *int     `json:"next,omitempty"`
	Prev       *int     `json:"prev,omitempty"`
	Data       []Worker `json:"data"`
}
type ProfileInput struct {
	Title          *string         `json:"title"`
	Bio            *string         `json:"bio"`
	HourlyRate     *float64        `json:"hourlyRate"`
	Availability   *string         `json:"availability"`
	Languages      json.RawMessage `json:"languages"`
	Portfolio      json.RawMessage `json:"portfolio"`
	Education      json.RawMessage `json:"education"`
	Certifications json.RawMessage `json:"certifications"`
}
type PublicProfile struct {
	Worker         map[string]any `json:"worker"`
	Stats          map[string]any `json:"stats"`
	Skills         []Skill        `json:"skills"`
	WorkHistory    []any          `json:"workHistory"`
	Portfolio      any            `json:"portfolio"`
	Languages      any            `json:"languages"`
	Availability   string         `json:"availability"`
	Education      any            `json:"education"`
	Certifications any            `json:"certifications"`
}
type Job struct {
	ID          string    `json:"_id"`
	JobTitle    string    `json:"jobTitle"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	MinBudget   float64   `json:"minBudget"`
	MaxBudget   float64   `json:"maxBudget"`
	WorkType    string    `json:"workType"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	ClientID    string    `json:"client"`
	CreatedAt   time.Time `json:"createdAt"`
}
type Offer struct {
	ID          string     `json:"_id"`
	JobID       string     `json:"jobId"`
	ClientID    string     `json:"clientId"`
	WorkerID    string     `json:"workerId"`
	Budget      float64    `json:"budget"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	Location    string     `json:"location,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
type OfferInput struct {
	JobID       string     `json:"jobId"`
	WorkerID    string     `json:"workerId"`
	Budget      float64    `json:"budget"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	Location    string     `json:"location"`
}
type OfferPage struct {
	Docs        []Offer `json:"docs"`
	TotalDocs   int64   `json:"totalDocs"`
	Limit       int     `json:"limit"`
	TotalPages  int     `json:"totalPages"`
	Page        int     `json:"page"`
	HasPrevPage bool    `json:"hasPrevPage"`
	HasNextPage bool    `json:"hasNextPage"`
	PrevPage    *int    `json:"prevPage"`
	NextPage    *int    `json:"nextPage"`
}
type OfferQuery struct {
	Page, Limit int
	Status      string
}
type JobContext struct {
	JobID           string         `json:"jobId"`
	Status          string         `json:"status"`
	Role            string         `json:"role"`
	Counterparty    map[string]any `json:"counterparty,omitempty"`
	CanReview       bool           `json:"canReview"`
	HasReviewed     bool           `json:"hasReviewed"`
	CanDispute      bool           `json:"canDispute"`
	HasOpenDispute  bool           `json:"hasOpenDispute"`
	ReviewEndpoint  string         `json:"reviewEndpoint,omitempty"`
	DisputeEndpoint string         `json:"disputeEndpoint,omitempty"`
}
type Review struct {
	ID        string    `json:"_id"`
	Job       any       `json:"job"`
	Reviewer  any       `json:"reviewer"`
	Reviewee  string    `json:"reviewee"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type ReviewResult struct {
	Reviews       []Review `json:"reviews"`
	TotalReviews  int      `json:"totalReviews"`
	AverageRating float64  `json:"averageRating"`
}
type Search struct {
	Page, Limit                                    int
	Search, SkillID, WorkType, Country, State, LGA string
	SavedBy                                        string
}
