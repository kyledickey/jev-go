package jev

import (
	"encoding/json"
	"fmt"
)

// DefaultModel is used when a request does not specify a model.
const DefaultModel = "jev-latest"

// Usage reports the token counts returned by the service.
type Usage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

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

// SystemOneResponse contains an evaluation's results and metadata.
type SystemOneResponse struct {
	// Model identifies the model that performed the evaluation.
	Model string `json:"model"`

	// Answers uses the same identifiers as the request's Questions map.
	Answers map[string]Answer `json:"answers"`

	// Usage contains the service-reported token counts.
	Usage Usage `json:"usage"`
}

// MarshalJSON encodes an evaluation request.
func (r SystemOneRequest) MarshalJSON() ([]byte, error) {
	if r.Model == "" {
		r.Model = DefaultModel
	}
	type wireRequest SystemOneRequest
	return json.Marshal(wireRequest(r))
}

// UnmarshalJSON decodes an evaluation response.
func (r *SystemOneResponse) UnmarshalJSON(data []byte) error {
	var wire struct {
		Model   string                     `json:"model"`
		Answers map[string]json.RawMessage `json:"answers"`
		Usage   json.RawMessage            `json:"usage"`
	}

	if err := unmarshalRequired(
		data, &wire, "model", "answers", "usage",
	); err != nil {
		return fmt.Errorf("jev: decode response: %w", err)
	}

	var usage Usage
	if err := unmarshalRequired(
		wire.Usage, &usage, "input_tokens", "output_tokens",
	); err != nil {
		return fmt.Errorf("jev: decode usage: %w", err)
	}

	answers := make(map[string]Answer, len(wire.Answers))

	for id, raw := range wire.Answers {
		answer, err := decodeAnswer(raw)
		if err != nil {
			return fmt.Errorf("jev: decode answer %q: %w", id, err)
		}
		answers[id] = answer
	}

	*r = SystemOneResponse{
		Model:   wire.Model,
		Answers: answers,
		Usage:   usage,
	}

	return nil
}
