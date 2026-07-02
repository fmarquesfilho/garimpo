package main

import (
	"testing"
)

func TestNewPipeline_RegistersReceivers(t *testing.T) {
	p := newTestPipeline(t)

	ids := p.ReceiverIDs()
	if len(ids) != 3 {
		t.Fatalf("expected 3 receivers, got %d: %v", len(ids), ids)
	}

	// Verifica que cada receiver do config foi registrado
	expected := map[string]bool{"shopee-products": false, "amazon-products": false, "shopee-coupons": false}
	for _, id := range ids {
		if _, ok := expected[id]; !ok {
			t.Errorf("unexpected receiver id: %s", id)
		}
		expected[id] = true
	}
	for id, found := range expected {
		if !found {
			t.Errorf("receiver %q não foi registrado", id)
		}
	}
}

func TestPipeline_GetProductSourceByMarketplace(t *testing.T) {
	p := newTestPipeline(t)

	// Shopee product source deve existir
	src, ok := p.GetProductSourceByMarketplace("shopee")
	if !ok {
		t.Fatal("expected shopee product source, got not found")
	}
	if src == nil {
		t.Fatal("expected non-nil source")
	}
	if src.Marketplace() != "shopee" {
		t.Errorf("marketplace = %q, want shopee", src.Marketplace())
	}

	// Amazon product source deve existir
	src, ok = p.GetProductSourceByMarketplace("amazon")
	if !ok {
		t.Fatal("expected amazon product source, got not found")
	}
	if src.Marketplace() != "amazon" {
		t.Errorf("marketplace = %q, want amazon", src.Marketplace())
	}

	// Mercado Livre não está configurado
	_, ok = p.GetProductSourceByMarketplace("mercadolivre")
	if ok {
		t.Error("expected mercadolivre not found, got found")
	}
}

func TestPipeline_GetCouponSourceByMarketplace(t *testing.T) {
	p := newTestPipeline(t)

	// Shopee coupon source deve existir
	src, ok := p.GetCouponSourceByMarketplace("shopee")
	if !ok {
		t.Fatal("expected shopee coupon source, got not found")
	}
	if src == nil {
		t.Fatal("expected non-nil coupon source")
	}
	if src.Marketplace() != "shopee" {
		t.Errorf("marketplace = %q, want shopee", src.Marketplace())
	}

	// Amazon coupon source NÃO está no testConfig
	_, ok = p.GetCouponSourceByMarketplace("amazon")
	if ok {
		t.Error("expected amazon coupon not found (not in test config), got found")
	}
}

func TestPipeline_GetProductSource_ByID(t *testing.T) {
	p := newTestPipeline(t)

	src, ok := p.GetProductSource("shopee-products")
	if !ok {
		t.Fatal("expected source by id, got not found")
	}
	if src == nil {
		t.Fatal("expected non-nil source")
	}

	// ID inexistente
	_, ok = p.GetProductSource("nao-existe")
	if ok {
		t.Error("expected not found for inexistent id")
	}
}

func TestPipeline_GetCouponSource_ByID(t *testing.T) {
	p := newTestPipeline(t)

	src, ok := p.GetCouponSource("shopee-coupons")
	if !ok {
		t.Fatal("expected coupon source by id, got not found")
	}
	if src == nil {
		t.Fatal("expected non-nil coupon source")
	}

	// Buscar product source por ID de coupon deve retornar not found
	_, ok = p.GetProductSource("shopee-coupons")
	if ok {
		t.Error("expected product source not found for coupon receiver")
	}
}

func TestNewPipeline_InvalidCron_ReturnsError(t *testing.T) {
	cfg := &CollectorConfig{
		Version: "1",
		Receivers: []ReceiverConfig{
			{
				ID:          "bad-cron",
				Type:        "product",
				Marketplace: "shopee",
				Schedule:    "not-a-cron-expression",
			},
		},
		Settings: SettingsConfig{MaxConcurrentReceivers: 1},
	}

	_, err := NewPipeline(cfg, testLogger())
	if err == nil {
		t.Fatal("expected error for invalid cron expression, got nil")
	}
}

func TestNewPipeline_UnknownType_ReturnsError(t *testing.T) {
	cfg := &CollectorConfig{
		Version: "1",
		Receivers: []ReceiverConfig{
			{
				ID:          "bad-type",
				Type:        "unknown",
				Marketplace: "shopee",
				Schedule:    "*/5 * * * *",
			},
		},
		Settings: SettingsConfig{MaxConcurrentReceivers: 1},
	}

	_, err := NewPipeline(cfg, testLogger())
	if err == nil {
		t.Fatal("expected error for unknown receiver type, got nil")
	}
}

func TestPipeline_StartStop(t *testing.T) {
	p := newTestPipeline(t)

	// Start não deve panicar
	p.Start()

	// Stop não deve panicar
	p.Stop(t.Context())
}
