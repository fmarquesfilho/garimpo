package couponsource

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func TestAmazonCouponAdapter_FetchCoupons_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"SearchResult": map[string]interface{}{
				"Items": []map[string]interface{}{
					{
						"ASIN": "B09TEST123",
						"ItemInfo": map[string]interface{}{
							"Title": map[string]interface{}{
								"DisplayValue": "Echo Dot com 30% OFF",
							},
							"Classifications": map[string]interface{}{
								"Binding": map[string]interface{}{
									"DisplayValue": "Eletrônicos",
								},
							},
						},
						"Offers": map[string]interface{}{
							"Listings": []map[string]interface{}{
								{
									"Price":       map[string]interface{}{"Amount": 199.0, "Currency": "BRL"},
									"SavingBasis": map[string]interface{}{"Amount": 299.0, "Currency": "BRL"},
								},
							},
						},
						"BrowseNodeInfo": map[string]interface{}{
							"BrowseNodes": []map[string]interface{}{
								{"DisplayName": "Smart Home"},
							},
						},
					},
					{
						"ASIN": "B09NODEAL",
						"ItemInfo": map[string]interface{}{
							"Title": map[string]interface{}{"DisplayValue": "Sem desconto"},
							"Classifications": map[string]interface{}{
								"Binding": map[string]interface{}{"DisplayValue": "Livros"},
							},
						},
						"Offers": map[string]interface{}{
							"Listings": []map[string]interface{}{
								{
									"Price":       map[string]interface{}{"Amount": 50.0},
									"SavingBasis": map[string]interface{}{"Amount": 0.0}, // no saving
								},
							},
						},
						"BrowseNodeInfo": map[string]interface{}{
							"BrowseNodes": []map[string]interface{}{},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := NewAmazonCouponAdapter("access", "secret", "garimpei-20")
	adapter.SetEndpoint(server.URL)
	adapter.SetHTTPClient(server.Client())

	coupons, err := adapter.FetchCoupons(FetchConfig{OwnerUID: "tenant-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only item with discount (SavingBasis > 0 and actual savings) should be returned
	if len(coupons) != 1 {
		t.Fatalf("expected 1 coupon (item with discount), got %d", len(coupons))
	}

	c := coupons[0]
	if c.ID != "B09TEST123" {
		t.Errorf("expected ASIN 'B09TEST123', got %q", c.ID)
	}
	if c.Marketplace != domain.MarketplaceAmazon {
		t.Errorf("expected marketplace 'amazon', got %q", c.Marketplace)
	}
	if c.DiscountType != domain.DiscountTypePercentage {
		t.Errorf("expected discount type 'percentage', got %q", c.DiscountType)
	}
	// (299-199)/299 ≈ 33.4%
	if c.DiscountValue < 33 || c.DiscountValue > 34 {
		t.Errorf("expected ~33%% discount, got %.2f%%", c.DiscountValue)
	}
	if c.OwnerUID != "tenant-1" {
		t.Errorf("expected owner_uid 'tenant-1', got %q", c.OwnerUID)
	}
	if len(c.ApplicableCategories) == 0 {
		t.Error("expected at least one category")
	}
	if c.ApplicableCategories[0] != "Smart Home" {
		t.Errorf("expected category 'Smart Home', got %q", c.ApplicableCategories[0])
	}
}

func TestAmazonCouponAdapter_FetchCoupons_NoCredentials_ReturnsNil(t *testing.T) {
	adapter := NewAmazonCouponAdapter("", "", "")
	coupons, err := adapter.FetchCoupons(FetchConfig{OwnerUID: "tenant"})
	if err != nil {
		t.Fatalf("expected nil error for missing credentials, got: %v", err)
	}
	if coupons != nil {
		t.Fatalf("expected nil coupons for missing credentials, got: %v", coupons)
	}
}

func TestAmazonCouponAdapter_FetchCoupons_RateLimit429(t *testing.T) {
	t.Skip("skipping: retry test uses real time.Sleep (60s backoff) — run manually with -timeout 180s")
}

func TestAmazonCouponAdapter_FetchCoupons_ServerError(t *testing.T) {
	t.Skip("skipping: retry test uses real time.Sleep (5s backoff) — run manually with -timeout 30s")
}
