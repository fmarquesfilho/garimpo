package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func testConfig() *CollectorConfig {
	return &CollectorConfig{
		Version: "1",
		Receivers: []ReceiverConfig{
			{
				ID:          "shopee-products",
				Type:        "product",
				Marketplace: "shopee",
				Schedule:    "*/30 * * * *",
			},
			{
				ID:          "amazon-products",
				Type:        "product",
				Marketplace: "amazon",
				Schedule:    "0 * * * *",
			},
			{
				ID:          "shopee-coupons",
				Type:        "coupon",
				Marketplace: "shopee",
				Schedule:    "0 */2 * * *",
			},
		},
		Settings: SettingsConfig{
			GRPCPort:               50051,
			HealthPort:             8081,
			MaxConcurrentReceivers: 3,
		},
	}
}

func newTestPipeline(t *testing.T) *Pipeline {
	t.Helper()
	cfg := testConfig()
	logger := testLogger()
	p, err := NewPipeline(cfg, logger)
	if err != nil {
		t.Fatalf("falha ao criar pipeline de teste: %v", err)
	}
	return p
}

func newTestCollectorServer(t *testing.T) *UnifiedCollectorServer {
	t.Helper()
	return NewUnifiedCollectorServer(newTestPipeline(t), store.NopSnapshots(), testLogger())
}

func newTestCouponServer(t *testing.T) *UnifiedCouponServer {
	t.Helper()
	return NewUnifiedCouponServer(newTestPipeline(t), testLogger())
}

func TestFetch_EmptyKeyword_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestCollectorServer(t)

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
	srv := newTestCollectorServer(t)

	_, err := srv.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "perfume",
		Limit:       5,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
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
	srv := newTestCollectorServer(t)

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

func TestCollect_EmptyBuscaId_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestCollectorServer(t)

	_, err := srv.Collect(context.Background(), &collectorpb.CollectRequest{
		Target:  &collectorpb.CollectRequest_Keyword{Keyword: "serum"},
		Limit:   10,
		BuscaId: "",
	})

	if err == nil {
		t.Fatal("expected error for empty busca_id, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v: %s", st.Code(), st.Message())
	}
}

func TestCollect_WithBuscaId_NoValidationError(t *testing.T) {
	srv := newTestCollectorServer(t)

	_, err := srv.Collect(context.Background(), &collectorpb.CollectRequest{
		Target:      &collectorpb.CollectRequest_Keyword{Keyword: "serum"},
		Limit:       10,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
		OwnerUid:    "user-123",
		BuscaId:     "busca-keyword-serum",
	})

	if err == nil {
		return // Success path if API is available
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	// Should NOT be InvalidArgument — busca_id is provided
	if st.Code() == codes.InvalidArgument {
		t.Errorf("busca_id was provided but got InvalidArgument: %s", st.Message())
	}
}

func TestFetch_UnconfiguredMarketplace_ReturnsUnimplemented(t *testing.T) {
	// Pipeline sem ML configurado → deve retornar Unimplemented
	srv := newTestCollectorServer(t)

	_, err := srv.Fetch(context.Background(), &collectorpb.FetchRequest{
		Keyword:     "perfume",
		Limit:       5,
		Marketplace: collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE,
	})

	if err == nil {
		t.Fatal("expected error for unconfigured marketplace, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("expected Unimplemented, got %v", st.Code())
	}
}

func TestFetchCoupons_UnconfiguredMarketplace_ReturnsUnimplemented(t *testing.T) {
	srv := newTestCouponServer(t)

	_, err := srv.FetchCoupons(context.Background(), &couponpb.FetchCouponsRequest{
		OwnerUid:    "test-owner",
		Marketplace: collectorpb.Marketplace_MARKETPLACE_AMAZON,
		PageSize:    100,
	})

	if err == nil {
		t.Fatal("expected error for unconfigured coupon marketplace, got nil")
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

// ─── Config tests ────────────────────────────────────────────────────────────

func TestLoadConfig_ValidYAML(t *testing.T) {
	// Cria um arquivo temporário de config
	tmpFile := t.TempDir() + "/collector.yaml"
	content := `
version: "1"
receivers:
  - id: test-products
    type: product
    marketplace: shopee
    schedule: "*/30 * * * *"
    credentials:
      app_id_env: TEST_APP_ID
      secret_env: TEST_SECRET
exporters:
  bigquery:
    project_env: BQ_PROJECT
    dataset: garimpo
    products_table: snapshots
    coupons_table: coupon_snapshots
settings:
  grpc_port: 50051
  health_port: 8081
  max_concurrent_receivers: 2
`
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Version != "1" {
		t.Errorf("version = %q, want %q", cfg.Version, "1")
	}
	if len(cfg.Receivers) != 1 {
		t.Errorf("receivers count = %d, want 1", len(cfg.Receivers))
	}
	if cfg.Receivers[0].ID != "test-products" {
		t.Errorf("receiver id = %q, want %q", cfg.Receivers[0].ID, "test-products")
	}
	if cfg.Settings.MaxConcurrentReceivers != 2 {
		t.Errorf("max_concurrent = %d, want 2", cfg.Settings.MaxConcurrentReceivers)
	}
}

func TestValidate_MissingVersion(t *testing.T) {
	cfg := &CollectorConfig{Receivers: []ReceiverConfig{{ID: "x", Type: "product", Marketplace: "shopee", Schedule: "* * * * *"}}}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for missing version")
	}
}

func TestValidate_DuplicateID(t *testing.T) {
	cfg := &CollectorConfig{
		Version: "1",
		Receivers: []ReceiverConfig{
			{ID: "dup", Type: "product", Marketplace: "shopee", Schedule: "* * * * *"},
			{ID: "dup", Type: "coupon", Marketplace: "shopee", Schedule: "* * * * *"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for duplicate ID")
	}
}

func TestValidate_InvalidType(t *testing.T) {
	cfg := &CollectorConfig{
		Version:   "1",
		Receivers: []ReceiverConfig{{ID: "x", Type: "invalid", Marketplace: "shopee", Schedule: "* * * * *"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid type")
	}
}
