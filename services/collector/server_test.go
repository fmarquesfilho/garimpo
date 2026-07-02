package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

func newTestServer() *CollectorServer {
	// Usa o ShopeeAdapter com credenciais fake — vai falhar na API real,
	// mas permite testar validação e lógica do server.
	src := source.NewShopeeAdapter("fake-app-id", "fake-secret")
	return NewCollectorServer(src)
}

func TestFetch_EmptyKeyword_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestServer()

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
	srv := newTestServer()

	_, err := srv.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword: "perfume",
		Limit:   5,
	})

	if err == nil {
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() == codes.InvalidArgument {
		t.Errorf("keyword 'perfume' should pass validation, got InvalidArgument: %v", st.Message())
	}
}

func TestFetchShop_ZeroShopId_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestServer()

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

func TestFetch_WrongMarketplace_ReturnsUnimplemented(t *testing.T) {
	srv := newTestServer()

	_, err := srv.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "perfume",
		Limit:       5,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_AMAZON,
	})

	if err == nil {
		t.Fatal("expected error for wrong marketplace, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("expected Unimplemented, got %v", st.Code())
	}
}

func TestParseNumericID(t *testing.T) {
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
			got := source.ParseNumericID(tc.input)
			if got != tc.want {
				t.Errorf("ParseNumericID(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
