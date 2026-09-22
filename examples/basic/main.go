// This example shows one complete Jev evaluation.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kyledickey/jev-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		var apiErr *jev.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "HTTP status: %d\n", apiErr.StatusCode)
		}

		os.Exit(1)
	}
}

func run() error {
	client, err := jev.NewClient()
	if err != nil {
		return fmt.Errorf("configure Jev: %w", err)
	}

	parent, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()

	response, err := client.SystemOne(ctx, jev.SystemOneRequest{
		State: "I bought this product 10 days ago and I want a full refund.",
		Questions: map[string]jev.Question{
			"urgent": jev.NoulQuestion{
				Instructions: "Is the customer eligible for a refund?",
				Criteria: &jev.NoulCriteria{
					True: "Refund Policy: full refunds eligible within 30 days of purchase.",
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("call Jev: %w", err)
	}

	answer, ok := response.Answers["urgent"]
	if !ok {
		return errors.New("response is missing urgent answer")
	}

	urgent, ok := answer.(*jev.NoulAnswer)
	if !ok {
		return fmt.Errorf("expected *jev.NoulAnswer for urgent got %T", answer)
	}

	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("Urgency probability: %.3f\n", urgent.Noul)
	fmt.Printf(
		"Tokens: %d input, %d output\n",
		response.Usage.InputTokens,
		response.Usage.OutputTokens,
	)

	return nil
}
