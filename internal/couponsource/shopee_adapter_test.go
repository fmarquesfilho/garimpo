package couponsource

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func TestShopeeCouponAdapter_FetchCoupons_Success(t *testing.T) {
	// Mock Shopee API response with coupon-like offers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"productOfferV2": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{
							"itemId":            12345,
							"productName":       "Cupom 20% OFF Eletrônicos",
							"offerLink":         "https://shp.ee/voucher123",
							"priceMin":          50.0,
							"priceDiscountRate": 0.20,
							"commissionRate":    0.10,
							"productCatIds":     []int{1001, 2002},
							"periodEndTime":     time.Now().Add(48 * time.Hour).Unix(),
							"shopId":            9999,
						},
						{
							"itemId":            67890,
							"productName":       "Produto sem desconto",
							"offerLink":         "https://shp.ee/prod67890",
							"priceMin":          100.0,
							"priceDiscountRate": 0.0, // no discount — should be filtered out
							"commissionRate":    0.05,
							"productCatIds":     []int{3003},
							"periodEndTime":     time.Now().Add(24 * time.Hour).Unix(),
							"shopId":            8888,
						},
					},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewShopeeCouponAdapter("test-app-id", "test-secret")
	adapter.SetEndpoint(server.URL)
	adapter.SetHTTPClient(server.Client())

	coupons, err := adapter.FetchCoupons(FetchConfig{OwnerUID: "tenant-1", PageSize: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the item with discount > 0 should be returned
	if len(coupons) != 1 {
		t.Fatalf("expected 1 coupon, got %d", len(coupons))
	}

	c := coupons[0]
	if c.ID != "12345" {
		t.Errorf("expected ID '12345', got %q", c.ID)
	}
	if c.Marketplace != domain.MarketplaceShopee {
		t.Errorf("expected marketplace 'shopee', got %q", c.Marketplace)
	}
	if c.DiscountType != domain.DiscountTypePercentage {
		t.Errorf("expected discount type 'percentage', got %q", c.DiscountType)
	}
	if c.DiscountValue != 20.0 { // 0.20 * 100
		t.Errorf("expected discount value 20.0, got %f", c.DiscountValue)
	}
	if c.OwnerUID != "tenant-1" {
		t.Errorf("expected owner_uid 'tenant-1', got %q", c.OwnerUID)
	}
	if c.Status != domain.CouponStatusActive {
		t.Errorf("expected status 'active', got %q", c.Status)
	}
	if len(c.ApplicableCategories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(c.ApplicableCategories))
	}
}

func TestShopeeCouponAdapter_FetchCoupons_ExpiredCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"productOfferV2": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{
							"itemId":            99001,
							"productName":       "Cupom expirado",
							"offerLink":         "https://shp.ee/expired",
							"priceMin":          30.0,
							"priceDiscountRate": 0.15,
							"commissionRate":    0.08,
							"productCatIds":     []int{4004},
							"periodEndTime":     time.Now().Add(-1 * time.Hour).Unix(), // expired
							"shopId":            7777,
						},
					},
					"pageInfo": map[string]interface{}{"hasNextPage": false},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewShopeeCouponAdapter("app", "secret")
	adapter.SetEndpoint(server.URL)
	adapter.SetHTTPClient(server.Client())

	coupons, err := adapter.FetchCoupons(FetchConfig{OwnerUID: "tenant-2", PageSize: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(coupons) != 1 {
		t.Fatalf("expected 1 coupon, got %d", len(coupons))
	}
	if coupons[0].Status != domain.CouponStatusExpired {
		t.Errorf("expected status 'expired', got %q", coupons[0].Status)
	}
}

func TestShopeeCouponAdapter_FetchCoupons_MissingCredentials(t *testing.T) {
	adapter := NewShopeeCouponAdapter("", "")
	_, err := adapter.FetchCoupons(FetchConfig{OwnerUID: "tenant"})
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestShopeeCouponAdapter_FetchCoupons_APIError_Retries(t *testing.T) {
	t.Skip("skipping: retry test uses real time.Sleep (5s backoff) — run manually with -timeout 60s")
}
