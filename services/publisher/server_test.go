package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	publisherpb "github.com/fmarquesfilho/garimpo/gen/go/publisher/v1"
)

func TestPublish_NilContent_ReturnsInvalidArgument(t *testing.T) {
	srv := NewPublisherServer()

	_, err := srv.Publish(context.Background(), &publisherpb.PublishRequest{
		Content: nil,
	})

	if err == nil {
		t.Fatal("expected error for nil content, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestListGroups_ReturnsEmpty(t *testing.T) {
	srv := NewPublisherServer()

	resp, err := srv.ListGroups(context.Background(), &publisherpb.ListGroupsRequest{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.GetGroups()) != 0 {
		t.Errorf("expected 0 groups, got %d", len(resp.GetGroups()))
	}
}
