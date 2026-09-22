package jev

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Answer represents one of the supported evaluation results.
type Answer interface {
	isAnswer()
}

// NoulAnswer contains the probability that a statement is true.
//
// see https://docs.typesafe.ai/primitives/noul
type NoulAnswer struct {
	// Noul ranges from 0 to 1
	Noul float64 `json:"noul"`
}

// Choice answer contains a selected option and its distribution.
//
// see https://docs.typesafe.ai/primitives/choice
type ChoiceAnswer struct {
	// Choice identifies the selected option by its original key.
	Choice string `json:"choice"`

	// Probabilities associates each option with its probability.
	Probabilities map[string]float64 `json:"probabilities"`

	// Confidence reports the service's confidence in the answer.
	Confidence float64 `json:"confidence"`
}

// ScoreAnswer contains a numeric score.
//
// see https://docs.typesafe.ai/primitives/score
type ScoreAnswer struct {
	// Score can fall between levels, so it is a float in the range of levels.
	Score float64 `json:"score"`

	// Legend associates level indicies with their descriptions.
	Legend map[string]string `json:"legend"`

	// Probabilities associates each level index with its probability.
	Probabilities map[string]float64 `json:"probabilities"`

	// Confidence is the model's confidence in the answer.
	Confidence float64 `json:"confidence"`
}

func (*NoulAnswer) isAnswer()   {}
func (*ChoiceAnswer) isAnswer() {}
func (*ScoreAnswer) isAnswer()  {}

var (
	_ Answer = (*NoulAnswer)(nil)
	_ Answer = (*ChoiceAnswer)(nil)
	_ Answer = (*ScoreAnswer)(nil)
)

// decodeAnswer selects an answer type using the JSON discriminator.
func decodeAnswer(data []byte) (Answer, error) {
	var header struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("decode answer type: %w", err)
	}

	var answer Answer
	var required []string

	switch header.Type {
	case "noul":
		answer = &NoulAnswer{}
		required = []string{"noul"}

	case "choice":
		answer = &ChoiceAnswer{}
		required = []string{"choice", "probabilities", "confidence"}

	case "score":
		answer = &ScoreAnswer{}
		required = []string{"score", "legend", "probabilities", "confidence"}

	default:
		return nil, fmt.Errorf("unsupported answer type %q", header.Type)
	}

	if err := unmarshalRequired(data, answer, required...); err != nil {
		return nil, fmt.Errorf("decode %s answer: %w", header.Type, err)
	}

	return answer, nil
}

// unmarshalRequired decodes an object after checking required properties.
func unmarshalRequired(data []byte, dst any, fields ...string) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}

	for _, field := range fields {
		value, ok := object[field]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("missing or null field %q", field)
		}
	}

	return json.Unmarshal(data, dst)
}
