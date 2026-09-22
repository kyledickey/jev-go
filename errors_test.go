package jev

import (
	"errors"
	"testing"
)

func TestAPIError(t *testing.T) {
	readErr := errors.New("boom")

	tests := []struct {
		name       string
		err        *APIError
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:    "status only",
			err:     &APIError{StatusCode: 404},
			wantMsg: "jev: API returned HTTP 404",
		},
		{
			name:       "with read error",
			err:        &APIError{StatusCode: 500, readErr: readErr},
			wantMsg:    "jev: API returned HTTP 500; failed to read response body",
			wantUnwrap: readErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
			if got := tt.err.Unwrap(); got != tt.wantUnwrap {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wantUnwrap)
			}
			if tt.wantUnwrap != nil && !errors.Is(tt.err, tt.wantUnwrap) {
				t.Error("errors.Is did not match wrapped read error")
			}
		})
	}
}
