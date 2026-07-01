package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
)

func TestFetch_EmptyKeyword_ReturnsInvalidArgument(t *testing.T) {
	srv := NewCollectorServer("fake-app-id", "fake-secret")

	_, err := srv.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword: "",
		Limit:   10,
	})

	if err == nil {
		t.Fatal("expected error for empty keyword, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestFetch_ValidKeyword_ReturnsNoError_OrInternalIfNoAPI(t *testing.T) {
	// With fake credentials, a real API call would fail with Internal error.
	// We only verify the validation passes (no InvalidArgument).
	srv := NewCollectorServer("fake-app-id", "fake-secret")

	_, err := srv.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword: "perfume",
		Limit:   5,
	})

	if err == nil {
		// If somehow the API responded, that's fine.
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	// Should NOT be InvalidArgument — it passed validation.
	if st.Code() == codes.InvalidArgument {
		t.Errorf("keyword 'perfume' should pass validation, got InvalidArgument: %v", st.Message())
	}
}

func TestFetchShop_ZeroShopId_ReturnsInvalidArgument(t *testing.T) {
	srv := NewCollectorServer("fake-app-id", "fake-secret")

	_, err := srv.FetchShop(context.Background(), &collectorpb.FetchShopRequest{
		ShopId: 0,
		Limit:  10,
	})

	if err == nil {
		t.Fatal("expected error for zero shop_id, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestParseItemID(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"12345", 12345},
		{"item-678", 678},
		{"", 0},
		{"abc", 0},
		{"shop_99_item_42", 9942},
		{"0", 0},
		{"1", 1},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseItemID(tc.input)
			if got != tc.want {
				t.Errorf("parseItemID(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
