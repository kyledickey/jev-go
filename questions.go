package jev

import "encoding/json"

// Question represents a question that can be included in an evaluation.
//
// A request can contain different question types in the same map.
// This interface lets each type keep its own fields.
type Question interface {
	json.Marshaler
	isQuestion()
}

// NoulQuestion asks for the probability that a statement is true.
//
// see https://docs.typesafe.ai/primitives/noul
type NoulQuestion struct {
	// Instructions describe what the model should evaluate.
	//
	// Supply a string for a simple question, or a JSON-compatible object
	// or array for structured instructions. For example, an object may
	// combine a question with additional context.
	Instructions any

	// Criteria optionally explain what true and false mean.
	Criteria *NoulCriteria
}

// NoulCriteria provides descriptions for the two possible outcomes.
//
// The descriptions may be strings, objects, or arrays, just like
// instructions.
type NoulCriteria struct {
	True  any `json:"true,omitempty"`
	False any `json:"false,omitempty"`
}

// isQuestion marks NoulQuestion as a question type
func (NoulQuestion) isQuestion() {}

// MarshalJSON implements json.Marshaler
func (q NoulQuestion) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type         string        `json:"type"`
		Instructions any           `json:"instructions"`
		Criteria     *NoulCriteria `json:"criteria,omitempty"`
	}{
		Type:         "noul",
		Instructions: q.Instructions,
		Criteria:     q.Criteria,
	})
}

var _ Question = NoulQuestion{}
