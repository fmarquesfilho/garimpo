package main

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cachepb "github.com/fmarquesfilho/garimpo/gen/go/cache/v1"
	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
)

// mockCollectorClient is a test double for CollectorServiceClient.
type mockCollectorClient struct {
	collectorpb.CollectorServiceClient

	fetchFunc     func(ctx context.Context, req *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error)
	fetchShopFunc func(ctx context.Context, req *collectorpb.FetchShopRequest) (*collectorpb.FetchShopResponse, error)
	fetchCount    atomic.Int64
}

func (m *mockCollectorClient) Fetch(ctx context.Context, req *collectorpb.FetchRequest, _ ...grpc.CallOption) (*collectorpb.FetchResponse, error) {
	m.fetchCount.Add(1)
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, req)
	}
	return &collectorpb.FetchResponse{
		Products: []*collectorpb.Product{
			{ItemId: 1, Name: "Serum Facial", Price: 29.90, ShopId: 100},
			{ItemId: 2, Name: "Serum Vitamina C", Price: 39.90, ShopId: 101},
		},
		TotalFound: 2,
		FetchedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

func (m *mockCollectorClient) FetchShop(ctx context.Context, req *collectorpb.FetchShopRequest, _ ...grpc.CallOption) (*collectorpb.FetchShopResponse, error) {
	m.fetchCount.Add(1)
	if m.fetchShopFunc != nil {
		return m.fetchShopFunc(ctx, req)
	}
	return &collectorpb.FetchShopResponse{
		Products: []*collectorpb.Product{
			{ItemId: 10, Name: "Produto Loja", Price: 19.90, ShopId: req.GetShopId()},
		},
		TotalFound: 1,
		FetchedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

func newTestServer(mock *mockCollectorClient) *CacheServer {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cache := NewLRUCache(256*1024*1024, 30*time.Minute)
	return NewCacheServer(cache, mock, 30*time.Minute, logger)
}

func TestGet_CacheHit(t *testing.T) {
	mock := &mockCollectorClient{}
	srv := newTestServer(mock)
	ctx := context.Background()

	// First call — cache miss, calls Collector
	resp, err := srv.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: []string{"serum"},
		BuscaId:        "busca-keyword-serum",
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:       "user1",
	})
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}
	if resp.CacheHit {
		t.Error("expected cache miss on first call")
	}
	if mock.fetchCount.Load() != 1 {
		t.Errorf("expected 1 Collector call, got %d", mock.fetchCount.Load())
	}

	// Second call — cache hit, no Collector call
	resp, err = srv.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: []string{"serum"},
		BuscaId:        "busca-keyword-serum",
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:       "user1",
	})
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	if !resp.CacheHit {
		t.Error("expected cache hit on second call")
	}
	if mock.fetchCount.Load() != 1 {
		t.Errorf("expected still 1 Collector call, got %d", mock.fetchCount.Load())
	}
	if len(resp.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(resp.Products))
	}
}

func TestGet_CacheMiss(t *testing.T) {
	mock := &mockCollectorClient{}
	srv := newTestServer(mock)
	ctx := context.Background()

	resp, err := srv.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: []string{"vitamina-c"},
		BuscaId:        "busca-keyword-vitc",
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:       "user1",
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if resp.CacheHit {
		t.Error("expected cache miss")
	}
	if mock.fetchCount.Load() != 1 {
		t.Errorf("expected 1 Collector call, got %d", mock.fetchCount.Load())
	}
	if resp.FetchedAt == "" {
		t.Error("expected fetched_at to be set")
	}
	if resp.SchemaVersion != schemaVersion {
		t.Errorf("expected schema_version %s, got %s", schemaVersion, resp.SchemaVersion)
	}
}

func TestGet_ValidationFailure(t *testing.T) {
	mock := &mockCollectorClient{
		fetchFunc: func(_ context.Context, _ *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error) {
			return nil, status.Error(codes.Internal, "marketplace unavailable")
		},
	}
	srv := newTestServer(mock)
	ctx := context.Background()

	_, err := srv.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: []string{"broken"},
		BuscaId:        "busca-broken",
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:       "user1",
	})
	if err == nil {
		t.Fatal("expected error from Collector failure")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal code, got %v", st.Code())
	}
}

func TestInvalidate_RemovesByBuscaId(t *testing.T) {
	mock := &mockCollectorClient{}
	srv := newTestServer(mock)
	ctx := context.Background()

	// Populate cache
	_, err := srv.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: []string{"serum"},
		BuscaId:        "busca-keyword-serum",
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:       "user1",
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Invalidate
	resp, err := srv.Invalidate(ctx, &cachepb.InvalidateRequest{
		BuscaId:  "busca-keyword-serum",
		OwnerUid: "user1",
	})
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.KeysRemoved != 1 {
		t.Errorf("expected 1 key removed, got %d", resp.KeysRemoved)
	}

	// Next Get should be a miss
	getResp, err := srv.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: []string{"serum"},
		BuscaId:        "busca-keyword-serum",
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:       "user1",
	})
	if err != nil {
		t.Fatalf("Get after invalidate failed: %v", err)
	}
	if getResp.CacheHit {
		t.Error("expected cache miss after invalidation")
	}
}

func TestInvalidate_Idempotent(t *testing.T) {
	mock := &mockCollectorClient{}
	srv := newTestServer(mock)
	ctx := context.Background()

	// Invalidate a non-existent busca — should succeed with 0 removed
	resp, err := srv.Invalidate(ctx, &cachepb.InvalidateRequest{
		BuscaId:  "nonexistent-busca",
		OwnerUid: "user1",
	})
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true (idempotent)")
	}
	if resp.KeysRemoved != 0 {
		t.Errorf("expected 0 keys removed, got %d", resp.KeysRemoved)
	}
}

func TestSingleflight_Coalesces(t *testing.T) {
	mock := &mockCollectorClient{
		fetchFunc: func(_ context.Context, _ *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error) {
			// Simulate slow Collector
			time.Sleep(50 * time.Millisecond)
			return &collectorpb.FetchResponse{
				Products: []*collectorpb.Product{
					{ItemId: 1, Name: "Slow Product", Price: 10.0},
				},
				TotalFound: 1,
				FetchedAt:  time.Now().Format(time.RFC3339),
			}, nil
		},
	}
	srv := newTestServer(mock)
	ctx := context.Background()

	// Fire 10 concurrent requests for the same key
	var wg sync.WaitGroup
	results := make([]*cachepb.GetResponse, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, err := srv.Get(ctx, &cachepb.GetRequest{
				CollectionKeys: []string{"concurrent-key"},
				BuscaId:        "busca-concurrent",
				Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
				OwnerUid:       "user1",
			})
			results[idx] = resp
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i, err := range errors {
		if err != nil {
			t.Errorf("request %d failed: %v", i, err)
		}
	}

	// Singleflight should coalesce: only 1 Collector call
	if mock.fetchCount.Load() != 1 {
		t.Errorf("expected 1 Collector call (singleflight), got %d", mock.fetchCount.Load())
	}
}

func TestGet_InvalidArguments(t *testing.T) {
	mock := &mockCollectorClient{}
	srv := newTestServer(mock)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *cachepb.GetRequest
	}{
		{
			name: "empty collection_keys",
			req:  &cachepb.GetRequest{BuscaId: "b1", OwnerUid: "u1"},
		},
		{
			name: "empty busca_id",
			req:  &cachepb.GetRequest{CollectionKeys: []string{"k1"}, OwnerUid: "u1"},
		},
		{
			name: "empty owner_uid",
			req:  &cachepb.GetRequest{CollectionKeys: []string{"k1"}, BuscaId: "b1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := srv.Get(ctx, tt.req)
			if err == nil {
				t.Error("expected error")
			}
			st, _ := status.FromError(err)
			if st.Code() != codes.InvalidArgument {
				t.Errorf("expected InvalidArgument, got %v", st.Code())
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	mock := &mockCollectorClient{}
	srv := newTestServer(mock)
	ctx := context.Background()

	resp, err := srv.Healthz(ctx, &cachepb.HealthzRequest{})
	if err != nil {
		t.Fatalf("Healthz failed: %v", err)
	}
	if !resp.Ready {
		t.Error("expected ready=true")
	}
}
