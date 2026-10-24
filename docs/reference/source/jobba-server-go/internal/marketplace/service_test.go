package marketplace

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	Repository
	recommendedLimit int
	toggle           bool
}

func (f *fakeRepo) Recommended(_ context.Context, _ string, n int) ([]Job, error) {
	f.recommendedLimit = n
	return []Job{}, nil
}
func (f *fakeRepo) ToggleSaved(context.Context, string, string) (bool, error) { return f.toggle, nil }
func TestOfferValidation(t *testing.T) {
	s := NewService(&fakeRepo{})
	_, e := s.CreateOffer(context.Background(), "client", OfferInput{Budget: -1})
	if !errors.Is(e, ErrInvalid) {
		t.Fatalf("expected invalid offer, got %v", e)
	}
}
func TestReviewValidation(t *testing.T) {
	s := NewService(&fakeRepo{})
	for _, rating := range []int{0, 6} {
		if _, e := s.CreateReview(context.Background(), "u", "j", rating, ""); !errors.Is(e, ErrInvalid) {
			t.Fatalf("rating %d: %v", rating, e)
		}
	}
}
func TestRecommendationLimitIsBounded(t *testing.T) {
	r := &fakeRepo{}
	s := NewService(r)
	_, _ = s.Recommended(context.Background(), "worker", 1000)
	if r.recommendedLimit != 100 {
		t.Fatalf("limit=%d", r.recommendedLimit)
	}
}
func TestCannotSaveSelf(t *testing.T) {
	s := NewService(&fakeRepo{})
	_, e := s.ToggleSaved(context.Background(), "same", "same")
	if !errors.Is(e, ErrWorkerNotFound) {
		t.Fatalf("got %v", e)
	}
}
