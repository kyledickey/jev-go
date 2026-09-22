package jev

import "encoding/json"

// DefaultModel is used when a request does not specify a model.
const DefaultModel = "jev-latest"

// SystemOneRequest contains the input for one evaluation.
type SystemOneRequest struct {
	// Model indicates the model to use.
	//
	// This defaults to "jev-latest".
	Model string `json:"model"`

	// State is the content evaluated by every question in this SystemOneRequest.
	//
	// Supply a string, a JSON-compatible object, or an array. Callers
	// can use their own structs; encoding/json will honor those structs'
	// exported fields and JSON tags.
	//
	// State is required.
	State any `json:"state"`

	// Questions maps caller-chosen identifiers to question definitions.
	//
	// The service returns answers under the same identifiers, callers can
	// associate each answer with its original question.
	Questions map[string]Question `json:"questions"`
}

// MarshalJSON encodes the request.
func (r SystemOneRequest) MarshalJSON() ([]byte, error) {
	if r.Model == "" {
		r.Model = DefaultModel
	}
	type wireRequest SystemOneRequest
	return json.Marshal(wireRequest(r))
}
