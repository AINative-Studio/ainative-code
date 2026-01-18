package backend

import "errors"

// Sentinel errors for common HTTP status codes
var (
	// ErrUnauthorized is returned when the server responds with 401 Unauthorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrPaymentRequired is returned when the server responds with 402 Payment Required
	ErrPaymentRequired = errors.New("payment required")

	// ErrServerError is returned when the server responds with 5xx status codes
	ErrServerError = errors.New("server error")

	// ErrBadRequest is returned when the server responds with 400 Bad Request
	ErrBadRequest = errors.New("bad request")

	// ErrNotFound is returned when the server responds with 404 Not Found
	ErrNotFound = errors.New("not found")
)
