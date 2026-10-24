package application

import "github.com/jobba/backend/internal/shared"

var (
	ErrUnauthorized        = &shared.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Authentication required"}
	ErrForbidden           = &shared.AppError{Code: "FORBIDDEN", Status: 403, Message: "You don't have permission to perform this action"}
	ErrInvalid             = &shared.AppError{Code: "INVALID_APPLICATION", Status: 400, Message: "Invalid application data"}
	ErrJobNotFound         = &shared.AppError{Code: "JOB_NOT_FOUND", Status: 404, Message: "Job not found"}
	ErrApplicationNotFound = &shared.AppError{Code: "APPLICATION_NOT_FOUND", Status: 404, Message: "Application not found"}
	ErrJobUnauthorized     = &shared.AppError{Code: "JOB_NOT_FOUND", Status: 404, Message: "Job not found or you are not authorized to view its applications"}
	ErrJobInactive         = &shared.AppError{Code: "JOB_NOT_ACTIVE", Status: 400, Message: "Job is not active"}
	ErrInvitationRequired  = &shared.AppError{Code: "JOB_PRIVATE", Status: 403, Message: "You don't have access to this job"}
	ErrOwnJob              = &shared.AppError{Code: "OWN_JOB_APPLICATION", Status: 400, Message: "You cannot apply to your own job"}
	ErrAlreadyApplied      = &shared.AppError{Code: "ALREADY_APPLIED", Status: 409, Message: "You have already applied for this job"}
	ErrAcceptForbidden     = &shared.AppError{Code: "APPLICATION_FORBIDDEN", Status: 403, Message: "Unauthorized to accept this application"}
	ErrOperation           = &shared.AppError{Code: "APPLICATION_OPERATION_FAILED", Status: 500, Message: "Application operation failed"}
)
