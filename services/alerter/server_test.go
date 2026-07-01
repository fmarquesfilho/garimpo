package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	alerterpb "github.com/fmarquesfilho/garimpo/gen/go/alerter/v1"
)

func TestCheckAndNotify_EmptyRules_ReturnsInvalidArgument(t *testing.T) {
	srv := NewAlerterServer()

	_, err := srv.CheckAndNotify(context.Background(), &alerterpb.CheckAndNotifyRequest{
		Rules: nil,
	})

	if err == nil {
		t.Fatal("expected error for empty rules, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}
