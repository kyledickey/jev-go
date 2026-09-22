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

// ChoiceQuestion asks the model to select one named option.
//
// see https://docs.typesafe.ai/primitives/choice
type ChoiceQuestion struct {
	// Instructions describe the decision to make.
	//
	// Supply a string or a JSON-compatible object or array.
	Instructions any

	// Criteria maps option identifiers to their descriptions.
	//
	// A description may be a string, object, array, or nil.
	//
	// The map is required. The API supports up to 255 options.
	Criteria map[string]any
}

// ScoreQuestion asks the model to evaluate the state against ordered levels.
//
// see https://docs.typesafe.ai/primitives/score
type ScoreQuestion struct {
	// Instructions describe what should be evaluated.
	//
	// Supply a string or JSON-compatible object or array.
	Instructions any

	// Criteria contains the level descriptions in their intended order.
	//
	// Each description may be a string, object, or array.
	// The API accepts between 2 and 10 levels.
	Criteria []any
}

func (NoulQuestion) isQuestion()   {}
func (ScoreQuestion) isQuestion()  {}
func (ChoiceQuestion) isQuestion() {}

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

// MarshalJSON implements json.Marshaler.
func (q ChoiceQuestion) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type         string         `json:"type"`
		Instructions any            `json:"instructions"`
		Criteria     map[string]any `json:"criteria"`
	}{
		Type:         "choice",
		Instructions: q.Instructions,
		Criteria:     q.Criteria,
	})
}

// MarshalJSON implements json.Marshaler.
func (q ScoreQuestion) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type         string `json:"type"`
		Instructions any    `json:"instructions"`
		Criteria     []any  `json:"criteria"`
	}{
		Type:         "score",
		Instructions: q.Instructions,
		Criteria:     q.Criteria,
	})
}

var (
	_ Question = NoulQuestion{}
	_ Question = ChoiceQuestion{}
	_ Question = ScoreQuestion{}
)
