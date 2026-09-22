package jev

import (
	"fmt"
	"net/http"
)

// APIError represents a non-successful HTTP reponse from the API.
//
// Transport failures that happen before the API response are not
// API errors.
type APIError struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	readErr    error
}

// Error returns a short description of the error.
func (e *APIError) Error() string {
	msg := fmt.Sprintf("jev: API returned HTTP %d", e.StatusCode)
	if e.readErr != nil {
		msg += "; failed to read response body"
	}
	return msg
}

func (e *APIError) Unwrap() error {
	return e.readErr
}
