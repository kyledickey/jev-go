package jev

import (
	"encoding/json"
	"testing"
)

func TestQuestionMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		q    Question
		want string
	}{
		{
			name: "noul with criteria",
			q: NoulQuestion{
				Instructions: "Is it urgent?",
				Criteria:     &NoulCriteria{True: "yes", False: "no"},
			},
			want: `{"type":"noul","instructions":"Is it urgent?","criteria":{"true":"yes","false":"no"}}`,
		},
		{
			name: "noul without criteria",
			q:    NoulQuestion{Instructions: "Is it urgent?"},
			want: `{"type":"noul","instructions":"Is it urgent?"}`,
		},
		{
			name: "noul partial criteria omits nil side",
			q: NoulQuestion{
				Instructions: "x",
				Criteria:     &NoulCriteria{True: "only true"},
			},
			want: `{"type":"noul","instructions":"x","criteria":{"true":"only true"}}`,
		},
		{
			name: "noul empty criteria struct",
			q:    NoulQuestion{Instructions: "x", Criteria: &NoulCriteria{}},
			want: `{"type":"noul","instructions":"x","criteria":{}}`,
		},
		{
			name: "noul structured instructions",
			q: NoulQuestion{
				Instructions: map[string]any{"text": "hi", "weight": 2},
			},
			want: `{"type":"noul","instructions":{"text":"hi","weight":2}}`,
		},
		{
			name: "zero noul",
			q:    NoulQuestion{},
			want: `{"type":"noul","instructions":null}`,
		},
		{
			name: "choice",
			q: ChoiceQuestion{
				Instructions: "Pick one",
				Criteria:     map[string]any{"a": "first", "b": []string{"x", "y"}},
			},
			want: `{"type":"choice","instructions":"Pick one","criteria":{"a":"first","b":["x","y"]}}`,
		},
		{
			name: "choice nil criteria",
			q:    ChoiceQuestion{Instructions: "Pick one"},
			want: `{"type":"choice","instructions":"Pick one","criteria":null}`,
		},
		{
			name: "choice empty criteria",
			q:    ChoiceQuestion{Instructions: "Pick one", Criteria: map[string]any{}},
			want: `{"type":"choice","instructions":"Pick one","criteria":{}}`,
		},
		{
			name: "score",
			q: ScoreQuestion{
				Instructions: "Rate it",
				Criteria:     []any{"bad", "ok", map[string]any{"label": "great"}},
			},
			want: `{"type":"score","instructions":"Rate it","criteria":["bad","ok",{"label":"great"}]}`,
		},
		{
			name: "score nil criteria",
			q:    ScoreQuestion{Instructions: "Rate it"},
			want: `{"type":"score","instructions":"Rate it","criteria":null}`,
		},
		{
			name: "score empty criteria",
			q:    ScoreQuestion{Instructions: "Rate it", Criteria: []any{}},
			want: `{"type":"score","instructions":"Rate it","criteria":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.q)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestQuestionMarshalJSON_PointerReceivers(t *testing.T) {
	got, err := json.Marshal(&ScoreQuestion{Instructions: "x", Criteria: []any{1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := `{"type":"score","instructions":"x","criteria":[1]}`; string(got) != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestQuestionMarshalJSON_UnmarshalableInstructions(t *testing.T) {
	bad := make(chan int)

	tests := []struct {
		name string
		q    Question
	}{
		{"noul", NoulQuestion{Instructions: bad}},
		{"choice", ChoiceQuestion{Instructions: bad}},
		{"score", ScoreQuestion{Instructions: bad}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := json.Marshal(tt.q); err == nil {
				t.Fatal("expected error for unmarshalable instructions")
			}
		})
	}
}

func TestQuestionMarshalJSON_InsideMap(t *testing.T) {
	questions := map[string]Question{
		"a": NoulQuestion{Instructions: "one"},
		"b": ChoiceQuestion{Instructions: "two", Criteria: map[string]any{"k": "v"}},
	}

	got, err := json.Marshal(questions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"a":{"type":"noul","instructions":"one"},"b":{"type":"choice","instructions":"two","criteria":{"k":"v"}}}`
	if string(got) != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
