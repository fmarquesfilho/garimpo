package couponsource

import (
	"errors"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func TestRegistry_Create_Shopee(t *testing.T) {
	src, err := DefaultRegistry.Create(domain.MarketplaceShopee, SourceConfig{
		AppID:  "test-app",
		Secret: "test-secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Marketplace() != domain.MarketplaceShopee {
		t.Errorf("expected marketplace 'shopee', got %q", src.Marketplace())
	}
	if src.Name() != "shopee-coupon-adapter" {
		t.Errorf("expected name 'shopee-coupon-adapter', got %q", src.Name())
	}
}

func TestRegistry_Create_Amazon(t *testing.T) {
	src, err := DefaultRegistry.Create(domain.MarketplaceAmazon, SourceConfig{
		AccessKey:  "key",
		SecretKey:  "secret",
		PartnerTag: "tag",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Marketplace() != domain.MarketplaceAmazon {
		t.Errorf("expected marketplace 'amazon', got %q", src.Marketplace())
	}
}

func TestRegistry_Create_Unsupported(t *testing.T) {
	_, err := DefaultRegistry.Create("aliexpress", SourceConfig{})
	if err == nil {
		t.Fatal("expected error for unsupported marketplace")
	}
	if !errors.Is(err, ErrUnsupportedMarketplace) {
		t.Errorf("expected ErrUnsupportedMarketplace, got: %v", err)
	}
}

func TestRegistry_Supported(t *testing.T) {
	supported := DefaultRegistry.Supported()
	if len(supported) < 2 {
		t.Fatalf("expected at least 2 supported marketplaces, got %d", len(supported))
	}

	has := func(s string) bool {
		for _, v := range supported {
			if v == s {
				return true
			}
		}
		return false
	}

	if !has(domain.MarketplaceShopee) {
		t.Error("expected shopee in supported list")
	}
	if !has(domain.MarketplaceAmazon) {
		t.Error("expected amazon in supported list")
	}
}
