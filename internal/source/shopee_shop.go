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
	StartPage     int           // página inicial (rotação); 0 ou 1 = primeira página
	PageDelay     time.Duration // delay entre páginas (throttling); 0 = sem delay
	ShopDelay     time.Duration // delay entre lojas (throttling); 0 = sem delay

	Endpoint   string
	HTTPClient *http.Client

	// LastPageInfo é preenchido após Fetch() com info de paginação por loja.
	// Chave: shopID, Valor: próxima página a buscar (para rotação).
	LastPageInfo map[int64]PageResult
}

// PageResult guarda o resultado de paginação para uma loja após Fetch.
type PageResult struct {
	NextPage    int  // próxima página a buscar na próxima coleta
	HasMore     bool // true se ainda há mais páginas no catálogo
	PagesFetched int // quantas páginas foram buscadas neste ciclo
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
		`{ productOfferV2(%s) { nodes { itemId productName productLink offerLink priceMin sales ratingStar commissionRate imageUrl shopName productCatIds shopId } pageInfo { page hasNextPage } } }`,
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
	startPage := s.StartPage
	if startPage <= 0 {
		startPage = 1
	}

	s.LastPageInfo = make(map[int64]PageResult)

	var produtos []domain.Product
	for i, shopID := range s.ShopIDs {
		// Throttling: delay entre lojas (exceto antes da primeira)
		if i > 0 && s.ShopDelay > 0 {
			time.Sleep(s.ShopDelay)
		}

		pagesFetched := 0
		hasMore := false
		nextPage := startPage

		for page := startPage; page < startPage+maxPages; page++ {
			// Throttling: delay entre páginas (exceto antes da primeira)
			if page > startPage && s.PageDelay > 0 {
				time.Sleep(s.PageDelay)
			}

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

			for _, n := range gql.Data.ProductOfferV2.Nodes {
				produtos = append(produtos, domain.Product{
					ID:         string(n.ItemID),
					Name:       n.ProductName,
					Category:   NomeCategoriaPrincipal(n.ProductCatIDs),
					Price:      float64(n.PriceMin),
					Commission: float64(n.CommissionRate),
					Sales30d:   int(n.Sales),
					Rating:     float64(n.RatingStar),
					Link:       n.OfferLink,
					Image:      n.ImageURL,
					ShopName:   n.ShopName,
					ShopID:     string(n.ShopID),
					CatIDs:     n.ProductCatIDs,
				})
			}

			pagesFetched++
			if !gql.Data.ProductOfferV2.PageInfo.HasNextPage {
				// Catálogo terminou — próxima coleta volta pra página 1
				nextPage = 1
				hasMore = false
				break
			}
			hasMore = true
			nextPage = page + 1
		}

		s.LastPageInfo[shopID] = PageResult{
			NextPage:     nextPage,
			HasMore:      hasMore,
			PagesFetched: pagesFetched,
		}
	}
	return produtos, nil
}

// shopGQLResponse é a resposta do productOfferV2 filtrado por shopId.
type shopGQLResponse struct {
	Data struct {
		ProductOfferV2 struct {
			Nodes    []productNode `json:"nodes"`
			PageInfo struct {
				Page        int  `json:"page"`
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"productOfferV2"`
	} `json:"data"`
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code int `json:"code"`
		} `json:"extensions"`
	} `json:"errors"`
}
