package jev

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeAnswer(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Answer
		wantErr string
	}{
		{
			name: "noul",
			data: `{"type":"noul","noul":0.75}`,
			want: &NoulAnswer{Noul: 0.75},
		},
		{
			name: "noul zero value present",
			data: `{"type":"noul","noul":0}`,
			want: &NoulAnswer{Noul: 0},
		},
		{
			name: "noul ignores unknown fields",
			data: `{"type":"noul","noul":0.5,"extra":true}`,
			want: &NoulAnswer{Noul: 0.5},
		},
		{
			name: "choice",
			data: `{"type":"choice","choice":"b","probabilities":{"a":0.2,"b":0.8},"confidence":0.9}`,
			want: &ChoiceAnswer{
				Choice:        "b",
				Probabilities: map[string]float64{"a": 0.2, "b": 0.8},
				Confidence:    0.9,
			},
		},
		{
			name: "choice empty probabilities",
			data: `{"type":"choice","choice":"a","probabilities":{},"confidence":1}`,
			want: &ChoiceAnswer{
				Choice:        "a",
				Probabilities: map[string]float64{},
				Confidence:    1,
			},
		},
		{
			name: "score",
			data: `{"type":"score","score":3,"legend":{"1":"bad","3":"good"},"probabilities":{"1":0.1,"3":0.9},"confidence":0.8}`,
			want: &ScoreAnswer{
				Score:         3,
				Legend:        map[string]string{"1": "bad", "3": "good"},
				Probabilities: map[string]float64{"1": 0.1, "3": 0.9},
				Confidence:    0.8,
			},
		},
		{
			name:    "noul missing value",
			data:    `{"type":"noul"}`,
			wantErr: `decode noul answer: missing or null field "noul"`,
		},
		{
			name:    "noul null value",
			data:    `{"type":"noul","noul":null}`,
			wantErr: `missing or null field "noul"`,
		},
		{
			name:    "noul wrong value type",
			data:    `{"type":"noul","noul":"high"}`,
			wantErr: "decode noul answer:",
		},
		{
			name:    "choice missing choice",
			data:    `{"type":"choice","probabilities":{},"confidence":1}`,
			wantErr: `missing or null field "choice"`,
		},
		{
			name:    "choice missing probabilities",
			data:    `{"type":"choice","choice":"a","confidence":1}`,
			wantErr: `missing or null field "probabilities"`,
		},
		{
			name:    "choice missing confidence",
			data:    `{"type":"choice","choice":"a","probabilities":{}}`,
			wantErr: `missing or null field "confidence"`,
		},
		{
			name:    "choice null probabilities",
			data:    `{"type":"choice","choice":"a","probabilities":null,"confidence":1}`,
			wantErr: `missing or null field "probabilities"`,
		},
		{
			name:    "score missing legend",
			data:    `{"type":"score","score":1,"probabilities":{},"confidence":1}`,
			wantErr: `missing or null field "legend"`,
		},
		{
			name:    "score missing score",
			data:    `{"type":"score","legend":{},"probabilities":{},"confidence":1}`,
			wantErr: `missing or null field "score"`,
		},
		{
			name:    "score wrong probabilities type",
			data:    `{"type":"score","score":1,"legend":{},"probabilities":[0.5],"confidence":1}`,
			wantErr: "decode score answer:",
		},
		{
			name:    "unsupported type",
			data:    `{"type":"ranking","ranking":[1,2]}`,
			wantErr: `unsupported answer type "ranking"`,
		},
		{
			name:    "missing type",
			data:    `{"noul":0.5}`,
			wantErr: `unsupported answer type ""`,
		},
		{
			name:    "null type",
			data:    `{"type":null,"noul":0.5}`,
			wantErr: `unsupported answer type ""`,
		},
		{
			name:    "type wrong json kind",
			data:    `{"type":7}`,
			wantErr: "decode answer type:",
		},
		{
			name:    "not an object",
			data:    `[1,2,3]`,
			wantErr: "decode answer type:",
		},
		{
			name:    "invalid json",
			data:    `{"type":"noul",`,
			wantErr: "decode answer type:",
		},
		{
			name:    "empty",
			data:    ``,
			wantErr: "decode answer type:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeAnswer([]byte(tt.data))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("got %#v, want error containing %q", got, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %q, want containing %q", err, tt.wantErr)
				}
				if got != nil {
					t.Errorf("answer = %#v, want nil on error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got  %#v\nwant %#v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalRequired(t *testing.T) {
	type target struct {
		A string `json:"a"`
		B int    `json:"b"`
	}

	tests := []struct {
		name    string
		data    string
		fields  []string
		want    target
		wantErr string
	}{
		{
			name:   "all present",
			data:   `{"a":"x","b":2}`,
			fields: []string{"a", "b"},
			want:   target{A: "x", B: 2},
		},
		{
			name:   "no required fields",
			data:   `{}`,
			fields: nil,
			want:   target{},
		},
		{
			name:   "extra fields ignored",
			data:   `{"a":"x","b":2,"c":true}`,
			fields: []string{"a"},
			want:   target{A: "x", B: 2},
		},
		{
			name:   "zero values count as present",
			data:   `{"a":"","b":0}`,
			fields: []string{"a", "b"},
			want:   target{},
		},
		{
			name:   "false and empty containers count as present",
			data:   `{"a":"x","b":1,"c":false,"d":[],"e":{}}`,
			fields: []string{"c", "d", "e"},
			want:   target{A: "x", B: 1},
		},
		{
			name:    "missing field",
			data:    `{"a":"x"}`,
			fields:  []string{"a", "b"},
			wantErr: `missing or null field "b"`,
		},
		{
			name:    "null field",
			data:    `{"a":null,"b":1}`,
			fields:  []string{"a"},
			wantErr: `missing or null field "a"`,
		},
		{
			name:    "null with surrounding whitespace",
			data:    "{\"a\": \n null \t,\"b\":1}",
			fields:  []string{"a"},
			wantErr: `missing or null field "a"`,
		},
		{
			name:    "first missing field reported",
			data:    `{}`,
			fields:  []string{"a", "b"},
			wantErr: `missing or null field "a"`,
		},
		{
			name:    "wrong field type",
			data:    `{"a":1,"b":2}`,
			fields:  []string{"a"},
			wantErr: "cannot unmarshal",
		},
		{
			name:    "not an object",
			data:    `"str"`,
			wantErr: "cannot unmarshal",
		},
		{
			name:    "json null",
			data:    `null`,
			fields:  []string{"a"},
			wantErr: `missing or null field "a"`,
		},
		{
			name:    "invalid json",
			data:    `{`,
			wantErr: "unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got target
			err := unmarshalRequired([]byte(tt.data), &got, tt.fields...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalRequired_ReturnsJSONErrors(t *testing.T) {
	var got struct{}
	err := unmarshalRequired([]byte(`{`), &got)
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("err = %T %v, want *json.SyntaxError", err, err)
	}
}
