package marketplace

import "context"

type Repository interface {
	SearchWorkers(context.Context, Search) (WorkerPage, error)
	Worker(context.Context, string) (Worker, error)
	ToggleSaved(context.Context, string, string) (bool, error)
	Profile(context.Context, string) (PublicProfile, error)
	UpsertProfile(context.Context, string, ProfileInput) (PublicProfile, error)
	Recommended(context.Context, string, int) ([]Job, error)
	CreateOffer(context.Context, string, OfferInput) (Offer, error)
	Offers(context.Context, string, string, OfferQuery) (OfferPage, error)
	Offer(context.Context, string, string, string) (Offer, error)
	AcceptOffer(context.Context, string, string) (Job, error)
	RejectOffer(context.Context, string, string) error
	CreateReview(context.Context, string, string, int, string) (Review, error)
	Reviews(context.Context, string) (ReviewResult, error)
	JobContext(context.Context, string, string, string) (JobContext, error)
}
type Service interface{ Repository }
