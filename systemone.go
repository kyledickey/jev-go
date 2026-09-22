package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	Model     string              `json:"model"` // Defaults to jev-latest
	State     any                 `json:"state"` // The content evaluated by the questions
	Questions map[string]Question `json:"questions"`
}

// SystemOneResponse contains an evaluation's results and metadata.
type SystemOneResponse struct {
	Model   string            `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage   Usage             `json:"usage"`
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

// SystemOne calls the API to evaluate the set of questions.
// Performs one attempt, no retry loop.
func (c *Client) SystemOne(ctx context.Context, req SystemOneRequest) (*SystemOneResponse, error) {
	if ctx == nil {
		return nil, errors.New("jev: context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("jev: evaluate: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("jev: encode request: %w", err)
	}

	ep := c.baseURL.JoinPath("systemone")
	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, ep.String(), bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("jev: create requests: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("jev: send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, readErr := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode < http.StatusOK ||
		httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{
			StatusCode: httpResp.StatusCode,
			Header:     httpResp.Header.Clone(),
			Body:       respBody,
			readErr:    readErr,
		}
	}

	if readErr != nil {
		return nil, fmt.Errorf("jev: read response: %w", readErr)
	}

	var resp SystemOneResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("jev: decode evaluation response: %w", err)
	}

	return &resp, nil
}
