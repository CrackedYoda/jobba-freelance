package chat

import "github.com/jobba/backend/internal/shared"

var (
	ErrUnauthorized    = &shared.AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Authentication required"}
	ErrForbidden       = &shared.AppError{Code: "FORBIDDEN", Status: 403, Message: "You don't have permission to perform this action"}
	ErrInvalid         = &shared.AppError{Code: "INVALID_CHAT_REQUEST", Status: 400, Message: "Invalid chat request"}
	ErrNotFound        = &shared.AppError{Code: "CONVERSATION_NOT_FOUND", Status: 404, Message: "Conversation not found or you are not a participant"}
	ErrMessageNotFound = &shared.AppError{Code: "MESSAGE_NOT_FOUND", Status: 404, Message: "Message not found or you are not the sender"}
	ErrParticipant     = &shared.AppError{Code: "INVALID_PARTICIPANT", Status: 400, Message: "Participant is not eligible for this conversation"}
	ErrOperation       = &shared.AppError{Code: "CHAT_OPERATION_FAILED", Status: 500, Message: "Chat operation failed"}
)
