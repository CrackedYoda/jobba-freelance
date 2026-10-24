package application

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}
func listInput(r *http.Request) (ListInput, error) {
	in := ListInput{Page: 1, Limit: 10}
	for key, target := range map[string]*int{"page": &in.Page, "limit": &in.Limit} {
		if raw := r.URL.Query().Get(key); raw != "" {
			v, e := strconv.Atoi(raw)
			if e != nil || v < 1 {
				return in, errors.New("page and limit must be positive integers; limit cannot exceed 100")
			}
			*target = v
		}
	}
	if in.Limit > 100 {
		return in, errors.New("page and limit must be positive integers; limit cannot exceed 100")
	}
	if raw := strings.ToUpper(r.URL.Query().Get("status")); raw != "" {
		s := Status(raw)
		if s != StatusPending && s != StatusShortlisted && s != StatusRejected && s != StatusAccepted {
			return in, errors.New("invalid application status")
		}
		in.Status = &s
	}
	return in, nil
}
func validID(id string) bool { return uuidPattern.MatchString(id) }
