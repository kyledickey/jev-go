package jev

// SystemOneRequest contains the input for one evaluation.
type SystemOneRequest struct {
	// Model indicates the model to use. Defaults to "jev-latest".
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
