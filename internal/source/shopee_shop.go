package source

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// ShopeeShopSource busca produtos de lojas específicas via shopOfferV2.
// Usa a mesma autenticação da ShopeeAPISource (AppID + Secret + HMAC-SHA256).
// Quando Keyword está preenchido, filtra os resultados da loja pelo termo.
type ShopeeShopSource struct {
	AppID  string
	Secret string

	ShopIDs       []int64 // IDs das lojas a monitorar
	Keyword       string  // filtro opcional dentro da loja
	CategoryLabel string
	Limit         int
	MaxPages      int

	Endpoint   string
	HTTPClient *http.Client
}

func NewShopeeShopSource(appID, secret string, shopIDs []int64) *ShopeeShopSource {
	return &ShopeeShopSource{
		AppID:    appID,
		Secret:   secret,
		ShopIDs:  shopIDs,
		Limit:    50,
		MaxPages: 2,
	}
}

func (s *ShopeeShopSource) Name() string { return "shopee-shop" }

func (s *ShopeeShopSource) buildQuery(shopID int64, page int) string {
	args := []string{
		fmt.Sprintf("shopId: %d", shopID),
		fmt.Sprintf("page: %d", page),
		fmt.Sprintf("limit: %d", s.Limit),
	}
	if s.Keyword != "" {
		args = append(args, fmt.Sprintf("keyword: %q", s.Keyword))
	}
	inner := strings.Join(args, ", ")
	return fmt.Sprintf(
		`{ shopOfferV2(%s) { nodes { itemId productName productLink offerLink priceMin sales ratingStar commissionRate shopName imageUrl } pageInfo { page hasNextPage } } }`,
		inner,
	)
}

func (s *ShopeeShopSource) endpoint() string {
	if s.Endpoint != "" {
		return s.Endpoint
	}
	return shopeeEndpoint
}

func (s *ShopeeShopSource) Fetch() ([]domain.Product, error) {
	if s.AppID == "" || s.Secret == "" {
		return nil, fmt.Errorf("shopee shop api: AppID/Secret não configurados")
	}
	if len(s.ShopIDs) == 0 {
		return nil, fmt.Errorf("shopee shop api: nenhum shopId configurado")
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	maxPages := s.MaxPages
	if maxPages <= 0 {
		maxPages = 1
	}

	var produtos []domain.Product
	for _, shopID := range s.ShopIDs {
		for page := 1; page <= maxPages; page++ {
			body, _ := json.Marshal(map[string]string{"query": s.buildQuery(shopID, page)})

			ts := strconv.FormatInt(time.Now().Unix(), 10)
			sum := sha256.Sum256([]byte(s.AppID + ts + string(body) + s.Secret))
			sig := hex.EncodeToString(sum[:])

			req, err := http.NewRequest(http.MethodPost, s.endpoint(), bytes.NewReader(body))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization",
				fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", s.AppID, ts, sig))

			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("shopee shop api: %w", err)
			}
			raw, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}

			var gql shopGQLResponse
			if err := json.Unmarshal(raw, &gql); err != nil {
				return nil, fmt.Errorf("shopee shop api: resposta inválida: %w", err)
			}
			if len(gql.Errors) > 0 {
				return nil, fmt.Errorf("shopee shop api: erro %d: %s",
					gql.Errors[0].Extensions.Code, gql.Errors[0].Message)
			}

			for _, n := range gql.Data.ShopOfferV2.Nodes {
				produtos = append(produtos, domain.Product{
					ID:         string(n.ItemID),
					Name:       n.ProductName,
					Category:   s.CategoryLabel,
					Price:      float64(n.PriceMin),
					Commission: float64(n.CommissionRate),
					Sales30d:   int(n.Sales),
					Rating:     float64(n.RatingStar),
					Link:       n.OfferLink,
					Image:      n.ImageURL,
				})
			}

			if !gql.Data.ShopOfferV2.PageInfo.HasNextPage {
				break
			}
		}
	}
	return produtos, nil
}

// shopGQLResponse é a resposta do shopOfferV2 (mesma estrutura de nodes).
type shopGQLResponse struct {
	Data struct {
		ShopOfferV2 struct {
			Nodes    []productNode `json:"nodes"`
			PageInfo struct {
				Page        int  `json:"page"`
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"shopOfferV2"`
	} `json:"data"`
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code int `json:"code"`
		} `json:"extensions"`
	} `json:"errors"`
}
