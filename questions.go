package jev

import "encoding/json"

// Question represents a question that can be included in an evaluation.
type Question interface {
	json.Marshaler
	isQuestion()
}

// NoulQuestion asks for the probability that a statement is true.
//
// see https://docs.typesafe.ai/primitives/noul
type NoulQuestion struct {
	Instructions any
	Criteria     *NoulCriteria
}

// NoulCriteria provides descriptions for the two possible outcomes.
type NoulCriteria struct {
	True  any `json:"true,omitempty"`
	False any `json:"false,omitempty"`
}

// ChoiceQuestion asks the model to select one named option.
//
// see https://docs.typesafe.ai/primitives/choice
type ChoiceQuestion struct {
	Instructions any
	Criteria     map[string]any
}

// ScoreQuestion asks the model to evaluate the state against ordered levels.
//
// see https://docs.typesafe.ai/primitives/score
type ScoreQuestion struct {
	Instructions any
	Criteria     []any
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
