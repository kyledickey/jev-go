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
	Noul float64 `json:"noul"` // 0-1
}

// Choice answer contains a selected option and its distribution.
//
// see https://docs.typesafe.ai/primitives/choice
type ChoiceAnswer struct {
	Choice        string             `json:"choice"` // The key of the selected option
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

// ScoreAnswer contains a numeric score.
//
// see https://docs.typesafe.ai/primitives/score
type ScoreAnswer struct {
	Score         float64            `json:"score"`
	Legend        map[string]string  `json:"legend"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
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
