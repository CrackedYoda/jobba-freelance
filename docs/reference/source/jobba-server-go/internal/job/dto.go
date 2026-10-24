package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type CreateInput struct {
	JobTitle      string        `json:"jobTitle"`
	Category      string        `json:"category"`
	Description   string        `json:"description"`
	SkillIDs      []string      `json:"skills"`
	Attachments   []string      `json:"attachments"`
	MinBudget     float64       `json:"minBudget"`
	MaxBudget     float64       `json:"maxBudget"`
	PaymentMethod PaymentMethod `json:"paymentMethod"`
	StartDate     time.Time     `json:"startDate"`
	EndDate       time.Time     `json:"endDate"`
	WorkType      WorkType      `json:"workType"`
	Location      string        `json:"location"`
	ApplicantType ApplicantType `json:"applicantType"`
	Privacy       Privacy       `json:"privacy"`
}

type UpdateInput struct {
	JobTitle      *string        `json:"jobTitle"`
	Category      *string        `json:"category"`
	Description   *string        `json:"description"`
	SkillIDs      *[]string      `json:"skills"`
	Attachments   *[]string      `json:"attachments"`
	MinBudget     *float64       `json:"minBudget"`
	MaxBudget     *float64       `json:"maxBudget"`
	PaymentMethod *PaymentMethod `json:"paymentMethod"`
	StartDate     *time.Time     `json:"startDate"`
	EndDate       *time.Time     `json:"endDate"`
	WorkType      *WorkType      `json:"workType"`
	Location      *string        `json:"location"`
	ApplicantType *ApplicantType `json:"applicantType"`
	Privacy       *Privacy       `json:"privacy"`
	Status        *Status        `json:"status"`
}

type assignInput struct {
	WorkerID string `json:"workerId"`
}

func (in UpdateInput) empty() bool {
	return in.JobTitle == nil && in.Category == nil && in.Description == nil && in.SkillIDs == nil && in.Attachments == nil && in.MinBudget == nil && in.MaxBudget == nil && in.PaymentMethod == nil && in.StartDate == nil && in.EndDate == nil && in.WorkType == nil && in.Location == nil && in.ApplicantType == nil && in.Privacy == nil && in.Status == nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func paginationFromRequest(r *http.Request) (Pagination, error) {
	p := Pagination{Page: 1, Limit: 10}
	for name, target := range map[string]*int{"page": &p.Page, "limit": &p.Limit} {
		if raw := r.URL.Query().Get(name); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 1 {
				return Pagination{}, errors.New("page and limit must be positive integers; limit cannot exceed 100")
			}
			*target = value
		}
	}
	if p.Limit > 100 || p.Page > 1_000_000 {
		return Pagination{}, errors.New("page and limit must be positive integers; limit cannot exceed 100")
	}
	return p, nil
}

func listFilterFromRequest(r *http.Request) (ListFilter, error) {
	p, err := paginationFromRequest(r)
	if err != nil {
		return ListFilter{}, err
	}
	q := r.URL.Query()
	f := ListFilter{Pagination: p, Category: strings.TrimSpace(q.Get("category")), Search: strings.TrimSpace(q.Get("search")), SortBy: q.Get("sortBy"), Order: q.Get("order")}
	if f.SortBy == "" {
		f.SortBy = "createdAt"
	}
	if f.Order == "" {
		f.Order = "desc"
	}
	if f.SortBy != "createdAt" && f.SortBy != "minBudget" && f.SortBy != "maxBudget" && f.SortBy != "startDate" {
		return f, errors.New("sortBy must be createdAt, minBudget, maxBudget, or startDate")
	}
	if f.Order != "asc" && f.Order != "desc" {
		return f, errors.New("order must be asc or desc")
	}
	if raw := q.Get("status"); raw != "" {
		value := Status(strings.ToUpper(raw))
		if !validStatus(value) {
			return f, errors.New("invalid status")
		}
		f.Status = &value
	}
	if raw := q.Get("workType"); raw != "" {
		value := WorkType(raw)
		if value != WorkOnsite && value != WorkRemote {
			return f, errors.New("invalid workType")
		}
		f.WorkType = &value
	}
	for name, target := range map[string]**float64{"minBudget": &f.MinBudget, "maxBudget": &f.MaxBudget} {
		if raw := q.Get(name); raw != "" {
			value, e := strconv.ParseFloat(raw, 64)
			if e != nil || value < 0 {
				return f, fmt.Errorf("%s must be a non-negative number", name)
			}
			*target = &value
		}
	}
	f.SkillIDs = q["skills"]
	return f, nil
}

func validStatus(s Status) bool {
	switch s {
	case StatusDraft, StatusOpen, StatusDisputed, StatusActive, StatusCompleted, StatusCancelled, StatusClosed:
		return true
	}
	return false
}
func validID(id string) bool { return uuidPattern.MatchString(id) }
