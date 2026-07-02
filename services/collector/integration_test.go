package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	"github.com/fmarquesfilho/garimpo/internal/couponsource"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// ─── Mock adapters ───────────────────────────────────────────────────────────

type mockProductSource struct {
	marketplace string
	products    []domain.Product
}

func (m *mockProductSource) Search(q source.SearchQuery) ([]domain.Product, error) {
	if q.Keyword == "error" {
		return nil, fmt.Errorf("simulated API error")
	}
	results := append([]domain.Product{}, m.products...)
	if q.Limit > 0 && len(results) > q.Limit {
		results = results[:q.Limit]
	}
	return results, nil
}

func (m *mockProductSource) FetchShop(shopID string, limit int) ([]domain.Product, error) {
	if shopID == "0" {
		return nil, fmt.Errorf("shop not found")
	}
	return m.products, nil
}

func (m *mockProductSource) Marketplace() string { return m.marketplace }
func (m *mockProductSource) Name() string        { return "mock-" + m.marketplace }

type mockCouponSource struct {
	marketplace string
	coupons     []domain.Coupon
}

func (m *mockCouponSource) FetchCoupons(cfg couponsource.FetchConfig) ([]domain.Coupon, error) {
	if cfg.OwnerUID == "error" {
		return nil, fmt.Errorf("simulated coupon error")
	}
	return m.coupons, nil
}

func (m *mockCouponSource) Marketplace() string { return m.marketplace }
func (m *mockCouponSource) Name() string        { return "mock-coupon-" + m.marketplace }

// ─── Test fixtures ───────────────────────────────────────────────────────────

var testProducts = []domain.Product{
	{ID: "P1", Name: "Sérum Vitamina C", Category: "Cuidados com a Pele", Price: 89.9, Commission: 0.15, Sales30d: 200, Rating: 4.9, Marketplace: domain.MarketplaceShopee},
	{ID: "P2", Name: "Perfume Kenzo", Category: "Perfumaria", Price: 299.9, Commission: 0.08, Sales30d: 80, Rating: 4.6, Marketplace: domain.MarketplaceShopee},
	{ID: "P3", Name: "Tônico BHA", Category: "Cuidados com a Pele", Price: 59.9, Commission: 0.12, Sales30d: 500, Rating: 4.8, Marketplace: domain.MarketplaceShopee},
}

var testCoupons = []domain.Coupon{
	{ID: "C1", Code: "SAVE10", Marketplace: domain.MarketplaceShopee, DiscountType: domain.DiscountTypePercentage, DiscountValue: 10, Status: "active"},
	{ID: "C2", Code: "FLAT20", Marketplace: domain.MarketplaceShopee, DiscountType: domain.DiscountTypeFixed, DiscountValue: 20, MinSpend: 100, Status: "active"},
}

// startTestGRPCServer creates a real gRPC server with mock adapters and returns a client connection.
func startTestGRPCServer(t *testing.T) (*grpc.ClientConn, func()) {
	t.Helper()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Build pipeline with mock sources injected directly
	cfg := &CollectorConfig{
		Version: "1",
		Receivers: []ReceiverConfig{
			{ID: "mock-shopee-products", Type: "product", Marketplace: "shopee", Schedule: "0 0 1 1 *"},
			{ID: "mock-shopee-coupons", Type: "coupon", Marketplace: "shopee", Schedule: "0 0 1 1 *"},
		},
		Settings: SettingsConfig{MaxConcurrentReceivers: 2},
	}

	p, err := NewPipeline(cfg, logger)
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	// Inject mock sources (override what the registry created)
	p.mu.Lock()
	p.receivers["mock-shopee-products"].ProductSource = &mockProductSource{marketplace: "shopee", products: testProducts}
	p.receivers["mock-shopee-coupons"].CouponSource = &mockCouponSource{marketplace: "shopee", coupons: testCoupons}
	p.mu.Unlock()

	// Start gRPC server on random port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	collectorpb.RegisterCollectorServiceServer(srv, NewUnifiedCollectorServer(p, logger))
	couponpb.RegisterCouponCollectorServiceServer(srv, NewUnifiedCouponServer(p, logger))

	go func() {
		_ = srv.Serve(lis)
	}()

	// Connect client
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		srv.Stop()
		t.Fatalf("dial: %v", err)
	}

	cleanup := func() {
		conn.Close()
		srv.GracefulStop()
	}

	return conn, cleanup
}

// ─── Integration tests ───────────────────────────────────────────────────────

func TestIntegration_Fetch_ReturnsProducts(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := collectorpb.NewCollectorServiceClient(conn)

	resp, err := client.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "sérum",
		Limit:       10,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
	})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if len(resp.Products) == 0 {
		t.Fatal("expected products, got empty")
	}
	if resp.TotalFound == 0 {
		t.Error("expected TotalFound > 0")
	}
	if resp.FetchedAt == "" {
		t.Error("expected FetchedAt timestamp")
	}

	// Verifica campos do primeiro produto
	p := resp.Products[0]
	if p.Name == "" {
		t.Error("expected product name")
	}
	if p.Category == "" {
		t.Error("expected product category")
	}
	if p.Commission == 0 {
		t.Error("expected non-zero commission")
	}
}

func TestIntegration_Fetch_EmptyKeyword_InvalidArgument(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := collectorpb.NewCollectorServiceClient(conn)

	_, err := client.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword: "",
		Limit:   10,
	})

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestIntegration_Fetch_UnknownMarketplace_Unimplemented(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := collectorpb.NewCollectorServiceClient(conn)

	_, err := client.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "perfume",
		Limit:       5,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE,
	})

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unimplemented {
		t.Errorf("expected Unimplemented, got %v", err)
	}
}

func TestIntegration_Fetch_WithLimit(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := collectorpb.NewCollectorServiceClient(conn)

	resp, err := client.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "sérum",
		Limit:       2,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
	})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if len(resp.Products) > 2 {
		t.Errorf("expected max 2 products, got %d", len(resp.Products))
	}
}

func TestIntegration_FetchShop_ReturnsProducts(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := collectorpb.NewCollectorServiceClient(conn)

	resp, err := client.FetchShop(context.Background(), &collectorpb.FetchShopRequest{
		ShopId:      123456,
		Limit:       10,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
	})
	if err != nil {
		t.Fatalf("FetchShop: %v", err)
	}

	if len(resp.Products) == 0 {
		t.Fatal("expected products from shop")
	}
}

func TestIntegration_FetchCoupons_ReturnsCoupons(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := couponpb.NewCouponCollectorServiceClient(conn)

	resp, err := client.FetchCoupons(context.Background(), &couponpb.FetchCouponsRequest{
		OwnerUid:    "test-owner",
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		PageSize:    100,
	})
	if err != nil {
		t.Fatalf("FetchCoupons: %v", err)
	}

	if len(resp.Coupons) != 2 {
		t.Errorf("expected 2 coupons, got %d", len(resp.Coupons))
	}
	if resp.TotalFound != 2 {
		t.Errorf("expected TotalFound=2, got %d", resp.TotalFound)
	}

	// Verificar campos do cupom
	c := resp.Coupons[0]
	if c.CouponId == "" {
		t.Error("expected coupon id")
	}
	if c.Code == "" {
		t.Error("expected coupon code")
	}
}

func TestIntegration_FetchCoupons_UnconfiguredMarketplace(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := couponpb.NewCouponCollectorServiceClient(conn)

	_, err := client.FetchCoupons(context.Background(), &couponpb.FetchCouponsRequest{
		OwnerUid:    "test-owner",
		Marketplace: collectorpb.Marketplace_MARKETPLACE_AMAZON,
		PageSize:    100,
	})

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unimplemented {
		t.Errorf("expected Unimplemented, got %v", err)
	}
}

func TestIntegration_Fetch_DefaultMarketplace_IsShopee(t *testing.T) {
	conn, cleanup := startTestGRPCServer(t)
	defer cleanup()

	client := collectorpb.NewCollectorServiceClient(conn)

	// MARKETPLACE_UNSPECIFIED should default to Shopee
	resp, err := client.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "perfume",
		Limit:       5,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_UNSPECIFIED,
	})
	if err != nil {
		t.Fatalf("Fetch with unspecified marketplace: %v", err)
	}
	if len(resp.Products) == 0 {
		t.Error("expected products (default to shopee)")
	}
}
