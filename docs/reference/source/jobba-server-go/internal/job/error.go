package job

import "github.com/jobba/backend/internal/shared"

var (
	ErrNotFound         = &shared.AppError{Code: "JOB_NOT_FOUND", Status: 404, Message: "Job not found"}
	ErrUnauthorized     = &shared.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Authentication required"}
	ErrForbidden        = &shared.AppError{Code: "FORBIDDEN", Status: 403, Message: "You don't have permission to perform this action"}
	ErrInvalid          = &shared.AppError{Code: "INVALID_JOB", Status: 400, Message: "Invalid job data"}
	ErrImmutable        = &shared.AppError{Code: "JOB_NOT_EDITABLE", Status: 400, Message: "Cannot update job that is active or completed"}
	ErrDeleteActive     = &shared.AppError{Code: "JOB_ACTIVE", Status: 400, Message: "Cannot delete job that is active"}
	ErrAlreadyCompleted = &shared.AppError{Code: "JOB_COMPLETED", Status: 400, Message: "Job is already completed"}
	ErrNotActive        = &shared.AppError{Code: "JOB_NOT_ACTIVE", Status: 400, Message: "Only active jobs can be completed"}
	ErrNotOpen          = &shared.AppError{Code: "JOB_NOT_OPEN", Status: 400, Message: "Can only assign workers to open jobs"}
	ErrInvalidWorker    = &shared.AppError{Code: "INVALID_WORKER", Status: 400, Message: "Invalid worker"}
	ErrPrivate          = &shared.AppError{Code: "JOB_PRIVATE", Status: 403, Message: "You don't have access to this job"}
	ErrRepository       = &shared.AppError{Code: "JOB_OPERATION_FAILED", Status: 500, Message: "Job operation failed"}
)
