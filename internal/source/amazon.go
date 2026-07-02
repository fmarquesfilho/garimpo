package source

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// Amazon Creators API endpoint (Brazil marketplace).
const amazonEndpoint = "https://webservices.amazon.com.br/paapi5/searchitems"

// AmazonCreatorsSource busca produtos via Amazon Creators API (substituta da PA-API 5.0).
//
// Autenticação: AWS Signature V4 (HMAC-SHA256) com AccessKey + SecretKey.
// Rate limit: 1 request/s por padrão (burst até 1 req/s sustained).
//
// Campos mapeados para o domínio:
//
//	ASIN           -> ID
//	Title          -> Name
//	Price.Amount   -> Price
//	StarRating     -> Rating
//	URL + tag      -> Link (link de afiliado com PartnerTag)
//
// Comissão: não vem no payload. É determinada por tabela fixa de categorias
// (1-15% conforme Amazon Associates). Usamos a tabela interna em commissionTable.
type AmazonCreatorsSource struct {
	AccessKey  string
	SecretKey  string
	PartnerTag string

	Keyword    string
	Limit      int
	SortBy     string // relevance, price-asc, price-desc, review-rank
	Endpoint   string
	HTTPClient *http.Client
}

func NewAmazonCreatorsSource(accessKey, secretKey, partnerTag string) *AmazonCreatorsSource {
	return &AmazonCreatorsSource{
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		PartnerTag: partnerTag,
		Limit:      10, // Amazon default is 10 items per request (max)
		SortBy:     "Relevance",
	}
}

func (s *AmazonCreatorsSource) Name() string { return "amazon-creators" }

func (s *AmazonCreatorsSource) Fetch() ([]domain.Product, error) {
	if s.AccessKey == "" || s.SecretKey == "" {
		return nil, fmt.Errorf("amazon AccessKey/SecretKey não configurados")
	}
	if s.Keyword == "" {
		return nil, fmt.Errorf("amazon keyword é obrigatório")
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = amazonEndpoint
	}

	limit := s.Limit
	if limit <= 0 || limit > 10 {
		limit = 10 // Creators API max is 10 per request
	}

	payload := searchItemsRequest{
		Keywords:    s.Keyword,
		PartnerTag:  s.PartnerTag,
		PartnerType: "Associates",
		Marketplace: "www.amazon.com.br",
		ItemCount:   limit,
		SortBy:      s.SortBy,
		Resources: []string{
			"ItemInfo.Title",
			"ItemInfo.Classifications",
			"Offers.Listings.Price",
			"Offers.Listings.SavingBasis",
			"Images.Primary.Large",
			"BrowseNodeInfo.BrowseNodes",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("amazon marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("amazon criar request: %w", err)
	}

	// AWS Signature V4 headers
	now := time.Now().UTC()
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	req.Header.Set("X-Amz-Target", "com.amazon.paapi5.v1.ProductAdvertisingAPIv1.SearchItems")
	req.Header.Set("Content-Encoding", "amz-1.0")
	s.signRequest(req, body, now)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("amazon api: requisição falhou: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("amazon ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("amazon api: status %d: %.500s", resp.StatusCode, string(raw))
	}

	var result searchItemsResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("amazon api: resposta inválida: %w", err)
	}

	var produtos []domain.Product
	for _, item := range result.SearchResult.Items {
		p := domain.Product{
			ID:          item.ASIN,
			Name:        item.ItemInfo.Title.DisplayValue,
			Category:    extractCategory(item),
			Marketplace: domain.MarketplaceAmazon,
		}

		// Price
		if len(item.Offers.Listings) > 0 {
			listing := item.Offers.Listings[0]
			p.Price = listing.Price.Amount
			if listing.SavingBasis.Amount > 0 {
				p.PriceMax = listing.SavingBasis.Amount
				if p.PriceMax > 0 {
					p.DiscountRate = (p.PriceMax - p.Price) / p.PriceMax
				}
			}
		}

		// Rating (não disponível diretamente no SearchItems; seria necessário GetItems com CustomerReviews resource)
		// Deixamos 0 e o scoring trata isso gracefully.

		// Image
		if item.Images.Primary.Large.URL != "" {
			p.Image = item.Images.Primary.Large.URL
		}

		// Link de afiliado: URL do produto + ?tag=PARTNER_TAG
		p.ProductLink = fmt.Sprintf("https://www.amazon.com.br/dp/%s", item.ASIN)
		p.Link = fmt.Sprintf("https://www.amazon.com.br/dp/%s?tag=%s", item.ASIN, s.PartnerTag)

		// Comissão por categoria (tabela fixa Amazon Associates Brasil)
		p.Commission = commissionForCategory(p.Category)

		produtos = append(produtos, p)
	}

	return produtos, nil
}

// signRequest aplica AWS Signature V4 simplificado ao request.
func (s *AmazonCreatorsSource) signRequest(req *http.Request, payload []byte, now time.Time) {
	dateStamp := now.Format("20060102")
	region := "us-east-1" // PA-API uses us-east-1 even for Brazil marketplace
	service := "ProductAdvertisingAPI"

	// Canonical request components
	payloadHash := sha256Hex(payload)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	credential := fmt.Sprintf("%s/%s/%s/%s/aws4_request", s.AccessKey, dateStamp, region, service)
	signedHeaders := "content-encoding;content-type;host;x-amz-content-sha256;x-amz-date;x-amz-target"

	canonicalHeaders := fmt.Sprintf("content-encoding:%s\ncontent-type:%s\nhost:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\nx-amz-target:%s\n",
		req.Header.Get("Content-Encoding"),
		req.Header.Get("Content-Type"),
		req.URL.Host,
		payloadHash,
		req.Header.Get("X-Amz-Date"),
		req.Header.Get("X-Amz-Target"),
	)

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		"POST",
		req.URL.Path,
		"", // no query string
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s/%s/%s/aws4_request\n%s",
		now.Format("20060102T150405Z"),
		dateStamp, region, service,
		sha256Hex([]byte(canonicalRequest)),
	)

	// Derive signing key
	kDate := hmacSHA256([]byte("AWS4"+s.SecretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	signature := hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s",
		credential, signedHeaders, signature,
	))
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// --- Amazon API request/response types ---

type searchItemsRequest struct {
	Keywords    string   `json:"Keywords"`
	PartnerTag  string   `json:"PartnerTag"`
	PartnerType string   `json:"PartnerType"`
	Marketplace string   `json:"Marketplace"`
	ItemCount   int      `json:"ItemCount"`
	SortBy      string   `json:"SortBy,omitempty"`
	Resources   []string `json:"Resources"`
}

type searchItemsResponse struct {
	SearchResult struct {
		Items      []amazonItem `json:"Items"`
		TotalCount int          `json:"TotalResultCount"`
	} `json:"SearchResult"`
	Errors []struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	} `json:"Errors"`
}

type amazonItem struct {
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
				Amount   float64 `json:"Amount"`
				Currency string  `json:"Currency"`
			} `json:"Price"`
			SavingBasis struct {
				Amount   float64 `json:"Amount"`
				Currency string  `json:"Currency"`
			} `json:"SavingBasis"`
		} `json:"Listings"`
	} `json:"Offers"`
	Images struct {
		Primary struct {
			Large struct {
				URL string `json:"URL"`
			} `json:"Large"`
		} `json:"Primary"`
	} `json:"Images"`
	BrowseNodeInfo struct {
		BrowseNodes []struct {
			DisplayName string `json:"DisplayName"`
		} `json:"BrowseNodes"`
	} `json:"BrowseNodeInfo"`
}

func extractCategory(item amazonItem) string {
	// Try BrowseNodes first (more specific), then Classifications.Binding
	if len(item.BrowseNodeInfo.BrowseNodes) > 0 {
		return item.BrowseNodeInfo.BrowseNodes[0].DisplayName
	}
	return item.ItemInfo.Classifications.Binding.DisplayValue
}

// commissionForCategory retorna a comissão fixa do programa Amazon Associates Brasil
// por categoria. Valores aproximados baseados na tabela pública (pode variar).
func commissionForCategory(category string) float64 {
	cat := strings.ToLower(category)
	switch {
	case strings.Contains(cat, "moda") || strings.Contains(cat, "roupa") || strings.Contains(cat, "vestuário"):
		return 0.15
	case strings.Contains(cat, "beleza") || strings.Contains(cat, "saúde") || strings.Contains(cat, "cuidados"):
		return 0.10
	case strings.Contains(cat, "livro") || strings.Contains(cat, "kindle"):
		return 0.08
	case strings.Contains(cat, "eletrônico") || strings.Contains(cat, "celular") || strings.Contains(cat, "computador"):
		return 0.03
	case strings.Contains(cat, "casa") || strings.Contains(cat, "cozinha") || strings.Contains(cat, "jardim"):
		return 0.08
	case strings.Contains(cat, "brinquedo") || strings.Contains(cat, "jogo"):
		return 0.06
	case strings.Contains(cat, "esporte") || strings.Contains(cat, "fitness"):
		return 0.06
	case strings.Contains(cat, "alimento") || strings.Contains(cat, "mercado"):
		return 0.05
	default:
		return 0.05 // fallback: 5% genérico
	}
}
