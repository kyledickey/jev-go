# jev-go

[![Go Reference](https://pkg.go.dev/badge/github.com/kyledickey/jev-go.svg)](https://pkg.go.dev/github.com/kyledickey/jev-go)
[![CI](https://github.com/kyledickey/jev-go/actions/workflows/ci.yml/badge.svg)](https://github.com/kyledickey/jev-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Go SDK for [TypeSafe's Jev](https://docs.typesafe.ai/introduction) API.

## Install

```sh
go get github.com/kyledickey/jev-go
```

The package name is `jev`.

## Usage

The client reads your key from `TYPESAFE_API_KEY` by default.

```go
client, err := jev.NewClient()
if err != nil {
	log.Fatal(err)
}

resp, err := client.SystemOne(ctx, jev.SystemOneRequest{
	State: "I bought this product 10 days ago and I want a full refund.",
	Questions: map[string]jev.Question{
		"eligible": jev.NoulQuestion{
			Instructions: "Is the customer eligible for a refund?",
			Criteria: &jev.NoulCriteria{
				True: "Full refunds are allowed within 30 days of purchase.",
			},
		},
	},
})
if err != nil {
	log.Fatal(err)
}

if a, ok := resp.Answers["eligible"].(*jev.NoulAnswer); ok {
	fmt.Printf("eligible: %.2f\n", a.Noul)
}
```

There are three question types, each with a matching answer type:

| Question         | Answer          |
| ---------------- | --------------- |
| `NoulQuestion`   | `*NoulAnswer`   |
| `ChoiceQuestion` | `*ChoiceAnswer` |
| `ScoreQuestion`  | `*ScoreAnswer`  |

See the [Jev docs](https://docs.typesafe.ai/primitives/noul) for how each one works, and [`examples/basic`](examples/basic/main.go) for a full program.

## Options

```go
client, err := jev.NewClient(
	jev.WithAPIKey("..."),
	jev.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	jev.WithBaseURL("https://api.typesafe.ai/v1"),
)
```

## Errors

Non-2xx responses come back as `*jev.APIError`, which has the status code, headers, and raw body.

```go
var apiErr *jev.APIError
if errors.As(err, &apiErr) {
	log.Printf("status %d: %s", apiErr.StatusCode, apiErr.Body)
}
```

Requests aren't retried, so wrap calls yourself if you need that.

## Contributing

Issues and PRs are welcome. For anything bigger than a small fix, it's worth opening an issue first so we can talk it through.

Before sending a PR, make sure these pass (CI runs the same checks):

```sh
gofmt -l .
go vet ./...
go test -race ./...
```

## License

[MIT](LICENSE)
