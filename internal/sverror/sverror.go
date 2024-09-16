package sverror

import (
	"fmt"
)

type ErrorCode error

var (
	InternalError      ErrorCode = fmt.Errorf("internal.error")
	AlreadyExistsError ErrorCode = fmt.Errorf("already.exists.error")
)

type ServiceError struct {
	Code        ErrorCode `json:"code,omitempty"`
	Message     string    `json:"message,omitempty"`
	ParentError error     `json:"error,omitempty"`
}

func (se *ServiceError) Error() string {
	return fmt.Sprintf(`code: %s, message: %s, error: %s`,
		se.Code,
		se.Message,
		se.ParentError,
	)
}

func NewServiceError(
	code ErrorCode,
	message string,
	parentError error,
) *ServiceError {
	return &ServiceError{
		Code:        code,
		Message:     message,
		ParentError: parentError,
	}
}

func NewInternalError(message string, parentError error) *ServiceError {
	return NewServiceError(InternalError, message, parentError)
}

func NewAlreadyExistsError(message string, parentError error) *ServiceError {
	return NewServiceError(AlreadyExistsError, message, parentError)
}
