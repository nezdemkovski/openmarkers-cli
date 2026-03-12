package api

import (
	"errors"
	"fmt"
)

var (
	ErrAuthRequired = errors.New("authentication required")
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrValidation   = errors.New("validation error")
	ErrServer       = errors.New("server error")
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewAPIError(statusCode int, message string) *APIError {
	e := &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
	switch {
	case statusCode == 401:
		e.Code = "auth_required"
		e.Err = ErrAuthRequired
	case statusCode == 403:
		e.Code = "forbidden"
		e.Err = ErrForbidden
	case statusCode == 404:
		e.Code = "not_found"
		e.Err = ErrNotFound
	case statusCode == 400:
		e.Code = "validation_error"
		e.Err = ErrValidation
	case statusCode >= 500:
		e.Code = "server_error"
		e.Err = ErrServer
	default:
		e.Code = "error"
		e.Err = errors.New(message)
	}
	return e
}

func ExitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrAuthRequired) {
		return 3
	}
	if errors.Is(err, ErrNotFound) {
		return 4
	}
	if errors.Is(err, ErrServer) {
		return 5
	}
	return 1
}
