package couponsource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
	"github.com/fmarquesfilho/garimpo/internal/domain"
)

const amazonCouponEndpoint = "https://webservices.amazon.com.br/paapi5/searchitems"

// AmazonCouponAdapter implements CouponSource for Amazon marketplace.
type AmazonCouponAdapter struct {
	accessKey  string
	secretKey  string
	partnerTag string
	endpoint   string
	client     *http.Client
}

func NewAmazonCouponAdapter(accessKey, secretKey, partnerTag string) *AmazonCouponAdapter {
	return &AmazonCouponAdapter{
		accessKey:  accessKey,
		secretKey:  secretKey,
		partnerTag: partnerTag,
	}
}

func (a *AmazonCouponAdapter) Marketplace() string { return domain.MarketplaceAmazon }
func (a *AmazonCouponAdapter) Name() string        { return "amazon-coupon-adapter" }

// SetEndpoint allows overriding for testing.
func (a *AmazonCouponAdapter) SetEndpoint(url string) { a.endpoint = url }

// SetHTTPClient allows injecting a test client.
func (a *AmazonCouponAdapter) SetHTTPClient(c *http.Client) { a.client = c }

func (a *AmazonCouponAdapter) FetchCoupons(cfg FetchConfig) ([]domain.Coupon, error) {
	if a.accessKey == "" || a.secretKey == "" {
		// Skip silently if no credentials (R2-AC8)
		return nil, nil
	}

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	endpoint := a.endpoint
	if endpoint == "" {
		endpoint = amazonCouponEndpoint
	}

	// Rate limit: 1 req/s
	coupons, err := a.fetchOffers(client, endpoint)
	if err != nil {
		// Retry logic
		var lastErr error
		for retry := 0; retry < 2; retry++ {
			if isRateLimit(err) {
				time.Sleep(60 * time.Second) // HTTP 429 → wait 60s
			} else {
				time.Sleep(5 * time.Second) // 5xx/timeout → 5s backoff
			}
			coupons, lastErr = a.fetchOffers(client, endpoint)
			if lastErr == nil {
				err = nil
				break
			}
			err = lastErr
		}
		if err != nil {
			return nil, fmt.Errorf("amazon coupon falhou após retries: %w", err)
		}
	}

	now := time.Now()
	for i := range coupons {
		coupons[i].OwnerUID = cfg.OwnerUID
		coupons[i].CollectedAt = now.Unix()
		coupons[i].Marketplace = domain.MarketplaceAmazon
		if coupons[i].EndTime > 0 && coupons[i].EndTime < now.Unix() {
			coupons[i].Status = domain.CouponStatusExpired
		} else {
			coupons[i].Status = domain.CouponStatusActive
		}
	}

	return coupons, nil
}

func (a *AmazonCouponAdapter) fetchOffers(client *http.Client, endpoint string) ([]domain.Coupon, error) { //nolint:funlen // HTTP response parsing requires sequential steps
	payload := map[string]interface{}{
		"Keywords":    "coupon deal",
		"PartnerTag":  a.partnerTag,
		"PartnerType": "Associates",
		"Marketplace": "www.amazon.com.br",
		"ItemCount":   10,
		"Resources": []string{
			"ItemInfo.Title",
			"ItemInfo.Classifications",
			"Offers.Listings.Price",
			"Offers.Listings.SavingBasis",
			"Offers.Listings.Promotions",
			"BrowseNodeInfo.BrowseNodes",
		},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("amazon coupon request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("amazon coupon api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("amazon coupon: %w", apperr.ErrRateLimited)
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("amazon coupon status %d: %w", resp.StatusCode, apperr.ErrAmazonAPI)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("amazon coupon status %d: %w", resp.StatusCode, apperr.ErrAmazonAPI)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("amazon coupon read body: %w", err)
	}

	var result struct {
		SearchResult struct {
			Items []struct {
				ASIN     string `json:"ASIN"`
				ItemInfo struct {
					Title struct {
						DisplayValue string `json:"DisplayValue"`
					} `json:"Title"`
					Classifications struct {
						Binding struct {
							DisplayValue string `json:"DisplayValue"`
						} `json:"Binding"`
					} `json:"Classifications"`
				} `json:"ItemInfo"`
				Offers struct {
					Listings []struct {
						Price struct {
							Amount float64 `json:"Amount"`
						} `json:"Price"`
						SavingBasis struct {
							Amount float64 `json:"Amount"`
						} `json:"SavingBasis"`
						Promotions []struct {
							Amount    float64 `json:"Amount"`
							Type      string  `json:"Type"`
							DiscPct   float64 `json:"DiscountPercent"`
							StartDate string  `json:"StartDate"`
							EndDate   string  `json:"EndDate"`
						} `json:"Promotions"`
					} `json:"Listings"`
				} `json:"Offers"`
				BrowseNodeInfo struct {
					BrowseNodes []struct {
						DisplayName string `json:"DisplayName"`
					} `json:"BrowseNodes"`
				} `json:"BrowseNodeInfo"`
			} `json:"Items"`
		} `json:"SearchResult"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("amazon coupon parse: %w", err)
	}

	var coupons []domain.Coupon
	for _, item := range result.SearchResult.Items {
		if len(item.Offers.Listings) == 0 {
			continue
		}
		listing := item.Offers.Listings[0]

		// Calculate discount
		var discountValue float64
		var discountType string
		if listing.SavingBasis.Amount > 0 && listing.Price.Amount > 0 {
			discountValue = ((listing.SavingBasis.Amount - listing.Price.Amount) / listing.SavingBasis.Amount) * 100
			discountType = domain.DiscountTypePercentage
		}
		if discountValue <= 0 {
			continue // skip items without discount
		}

		categories := make([]string, 0)
		for _, bn := range item.BrowseNodeInfo.BrowseNodes {
			categories = append(categories, bn.DisplayName)
		}
		if len(categories) == 0 && item.ItemInfo.Classifications.Binding.DisplayValue != "" {
			categories = append(categories, item.ItemInfo.Classifications.Binding.DisplayValue)
		}

		link := fmt.Sprintf("https://www.amazon.com.br/dp/%s?tag=%s", item.ASIN, a.partnerTag)

		coupons = append(coupons, domain.Coupon{
			ID:                   item.ASIN,
			Code:                 link,
			DiscountType:         discountType,
			DiscountValue:        discountValue,
			ApplicableCategories: categories,
		})
	}

	return coupons, nil
}

func isRateLimit(err error) bool {
	return errors.Is(err, apperr.ErrRateLimited)
}
