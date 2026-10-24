package marketplace

import "github.com/jobba/backend/internal/shared"

var (
	ErrUnauthorized    = &shared.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Authentication required"}
	ErrForbidden       = &shared.AppError{Code: "FORBIDDEN", Status: 403, Message: "You don't have permission to perform this action"}
	ErrWorkerNotFound  = &shared.AppError{Code: "WORKER_NOT_FOUND", Status: 404, Message: "Worker not found or inactive"}
	ErrJobNotFound     = &shared.AppError{Code: "JOB_NOT_FOUND", Status: 404, Message: "Job not found or you are not the owner"}
	ErrOfferNotFound   = &shared.AppError{Code: "OFFER_NOT_FOUND", Status: 404, Message: "Offer not found or already processed"}
	ErrInvalid         = &shared.AppError{Code: "INVALID_MARKETPLACE_REQUEST", Status: 400, Message: "Invalid request data"}
	ErrJobState        = &shared.AppError{Code: "INVALID_JOB_STATE", Status: 400, Message: "Job is not available for this action"}
	ErrAlreadyReviewed = &shared.AppError{Code: "ALREADY_REVIEWED", Status: 400, Message: "You have already reviewed this job"}
	ErrReviewForbidden = &shared.AppError{Code: "REVIEW_FORBIDDEN", Status: 403, Message: "You are not authorized to review this job"}
	ErrOperation       = &shared.AppError{Code: "MARKETPLACE_OPERATION_FAILED", Status: 500, Message: "Marketplace operation failed"}
)
