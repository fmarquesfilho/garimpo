package couponsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
	"github.com/fmarquesfilho/garimpo/internal/domain"
)

const shopeeEndpoint = "https://open-api.affiliate.shopee.com.br/graphql"

// ShopeeCouponAdapter implements CouponSource for Shopee marketplace.
type ShopeeCouponAdapter struct {
	appID    string
	secret   string
	endpoint string
	client   *http.Client
}

func NewShopeeCouponAdapter(appID, secret string) *ShopeeCouponAdapter {
	return &ShopeeCouponAdapter{
		appID:  appID,
		secret: secret,
	}
}

func (a *ShopeeCouponAdapter) Marketplace() string { return domain.MarketplaceShopee }
func (a *ShopeeCouponAdapter) Name() string        { return "shopee-coupon-adapter" }

// SetEndpoint allows overriding the endpoint for testing.
func (a *ShopeeCouponAdapter) SetEndpoint(url string) { a.endpoint = url }

// SetHTTPClient allows injecting a test client.
func (a *ShopeeCouponAdapter) SetHTTPClient(c *http.Client) { a.client = c }

func (a *ShopeeCouponAdapter) FetchCoupons(cfg FetchConfig) ([]domain.Coupon, error) {
	if a.appID == "" || a.secret == "" {
		return nil, fmt.Errorf("shopee coupon: %w", apperr.ErrNoConfig)
	}

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	endpoint := a.endpoint
	if endpoint == "" {
		endpoint = shopeeEndpoint
	}
	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = 500
	}

	var coupons []domain.Coupon
	now := time.Now()

	for page := 1; ; page++ {
		// Throttle: 200ms between pages (except first)
		if page > 1 {
			time.Sleep(200 * time.Millisecond)
		}

		result, hasNext, err := a.fetchPage(client, endpoint, page, pageSize)
		if err != nil {
			// Retry up to 2 times with 5s backoff
			var lastErr error
			for retry := 0; retry < 2; retry++ {
				time.Sleep(5 * time.Second)
				result, hasNext, lastErr = a.fetchPage(client, endpoint, page, pageSize)
				if lastErr == nil {
					err = nil
					break
				}
			}
			if err != nil && lastErr != nil {
				return nil, fmt.Errorf("shopee coupon falhou após retries: %w", lastErr)
			}
		}

		for i := range result {
			result[i].OwnerUID = cfg.OwnerUID
			result[i].CollectedAt = now.Unix()
			result[i].Marketplace = domain.MarketplaceShopee
			// Mark expired
			if result[i].EndTime > 0 && result[i].EndTime < now.Unix() {
				result[i].Status = domain.CouponStatusExpired
			} else {
				result[i].Status = domain.CouponStatusActive
			}
		}
		coupons = append(coupons, result...)

		if !hasNext {
			break
		}
	}

	return coupons, nil
}

func (a *ShopeeCouponAdapter) fetchPage(client *http.Client, endpoint string, page, pageSize int) ([]domain.Coupon, bool, error) {
	query := fmt.Sprintf(
		`{ productOfferV2(listType: 3, sortType: 1, page: %d, limit: %d) { nodes { itemId productName offerLink priceMin priceDiscountRate commissionRate productCatIds periodEndTime shopId } pageInfo { page hasNextPage } } }`,
		page, pageSize,
	)

	body, _ := json.Marshal(map[string]string{"query": query})

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sum := sha256.Sum256([]byte(a.appID + ts + string(body) + a.secret))
	sig := hex.EncodeToString(sum[:])

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization",
		fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", a.appID, ts, sig))

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("shopee coupon api: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("ler resposta: %w", err)
	}

	var gql struct {
		Data struct {
			ProductOfferV2 struct {
				Nodes []struct {
					ItemID            json.Number `json:"itemId"`
					ProductName       string      `json:"productName"`
					OfferLink         string      `json:"offerLink"`
					PriceMin          float64     `json:"priceMin"`
					PriceDiscountRate float64     `json:"priceDiscountRate"`
					CommissionRate    float64     `json:"commissionRate"`
					ProductCatIDs     []int       `json:"productCatIds"`
					PeriodEndTime     int64       `json:"periodEndTime"`
					ShopID            json.Number `json:"shopId"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"productOfferV2"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(raw, &gql); err != nil {
		return nil, false, fmt.Errorf("parse resposta: %w", err)
	}
	if len(gql.Errors) > 0 {
		return nil, false, fmt.Errorf("shopee coupon %s: %w", gql.Errors[0].Message, apperr.ErrShopeeAPI)
	}

	var coupons []domain.Coupon
	for _, n := range gql.Data.ProductOfferV2.Nodes {
		// Only include items with discount (coupon-like offers)
		if n.PriceDiscountRate <= 0 {
			continue
		}

		categories := make([]string, len(n.ProductCatIDs))
		for i, c := range n.ProductCatIDs {
			categories[i] = strconv.Itoa(c)
		}

		coupons = append(coupons, domain.Coupon{
			ID:                   n.ItemID.String(),
			Code:                 n.OfferLink,
			DiscountType:         domain.DiscountTypePercentage,
			DiscountValue:        n.PriceDiscountRate * 100, // convert 0.20 → 20%
			MinSpend:             n.PriceMin,
			EndTime:              n.PeriodEndTime,
			ApplicableCategories: categories,
		})
	}

	return coupons, gql.Data.ProductOfferV2.PageInfo.HasNextPage, nil
}
