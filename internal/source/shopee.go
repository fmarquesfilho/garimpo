package source

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// Endpoint GraphQL de afiliados (Brasil). Para outros países, troque o TLD.
const shopeeEndpoint = "https://open-api.affiliate.shopee.com.br/graphql"

// ShopeeAPISource é o ADAPTADOR da API de afiliados da Shopee (Incremento 1).
// Implementa a mesma porta ProductSource: trocar CSVSource por este não altera
// nada no motor, nas estratégias ou no ranking.
//
// Autenticação (confirmada na doc da Shopee):
//
//	Authorization: SHA256 Credential={AppId}, Timestamp={ts}, Signature={sig}
//	Signature = SHA256(AppId + ts + Payload + Secret)
//
// onde Payload é o corpo JSON EXATO enviado. Por isso assinamos os mesmos bytes
// que vão no body — qualquer divergência de whitespace quebra a assinatura
// (erro 10020). Timestamp em segundos Unix; janela de ~5 min.
//
// Campos relevantes do productOfferV2 mapeados para o domínio:
//
//	commissionRate -> Commission   (já é fração, ex.: 0.0850)
//	priceMin       -> Price
//	sales          -> Sales30d     (ver nota abaixo)
//	ratingStar     -> Rating
//	offerLink      -> Link         (link já com seu tracking de afiliado)
//
// NOTA sobre `sales`: a API devolve o volume de vendas do produto como a Shopee
// o reporta (acumulado/período), não estritamente "últimos 30 dias". Para
// ranking serve bem como proxy de demanda — só não interprete como janela fixa.
type ShopeeAPISource struct {
	AppID  string
	Secret string

	ListType     int
	SortType     int
	ProductCatID int
	Keyword      string
	ItemID       string // busca por itemId específico (resolver-link)
	Limit        int
	MaxPages     int

	CategoryLabel string
	Endpoint      string
	HTTPClient    *http.Client
}

// NewShopeeAPISource traz padrões alinhados à regra dela (priorizar comissão).
func NewShopeeAPISource(appID, secret string) *ShopeeAPISource {
	return &ShopeeAPISource{
		AppID:    appID,
		Secret:   secret,
		ListType: 1, // Maior comissão
		SortType: 5, // Comissão (desc)
		Limit:    50,
		MaxPages: 1,
	}
}

func (s *ShopeeAPISource) Name() string { return "shopee-api" }

func (s *ShopeeAPISource) buildQuery(page int) string {
	args := []string{
		fmt.Sprintf("listType: %d", s.ListType),
		fmt.Sprintf("sortType: %d", s.SortType),
		fmt.Sprintf("page: %d", page),
		fmt.Sprintf("limit: %d", s.Limit),
	}
	if s.ProductCatID != 0 {
		args = append(args, fmt.Sprintf("productCatId: %d", s.ProductCatID))
	}
	if s.ItemID != "" {
		args = append(args, fmt.Sprintf("itemId: %s", s.ItemID))
	}
	if s.Keyword != "" {
		args = append(args, fmt.Sprintf("keyword: %q", s.Keyword))
	}
	inner := strings.Join(args, ", ")
	return fmt.Sprintf(
		`{ productOfferV2(%s) { nodes { itemId productName productLink offerLink priceMin sales ratingStar commissionRate shopName imageUrl } pageInfo { page hasNextPage } } }`,
		inner,
	)
}

type productNode struct {
	ItemID         flexString `json:"itemId"`
	ProductName    string     `json:"productName"`
	ProductLink    string     `json:"productLink"`
	OfferLink      string     `json:"offerLink"`
	PriceMin       flexFloat  `json:"priceMin"`
	Sales          flexInt    `json:"sales"`
	RatingStar     flexFloat  `json:"ratingStar"`
	CommissionRate flexFloat  `json:"commissionRate"`
	ShopName       string     `json:"shopName"`
	ImageURL       string     `json:"imageUrl"`
}

type gqlResponse struct {
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

func (s *ShopeeAPISource) Fetch() ([]domain.Product, error) {
	if s.AppID == "" || s.Secret == "" {
		return nil, errors.New("shopee api: AppID/Secret não configurados")
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	maxPages := s.MaxPages
	if maxPages <= 0 {
		maxPages = 1
	}
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = shopeeEndpoint
	}

	var produtos []domain.Product
	for page := 1; page <= maxPages; page++ {
		body, err := json.Marshal(map[string]string{"query": s.buildQuery(page)})
		if err != nil {
			return nil, err
		}

		ts := strconv.FormatInt(time.Now().Unix(), 10)
		sum := sha256.Sum256([]byte(s.AppID + ts + string(body) + s.Secret))
		sig := hex.EncodeToString(sum[:])

		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization",
			fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", s.AppID, ts, sig))

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("shopee api: requisição falhou: %w", err)
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var gql gqlResponse
		if err := json.Unmarshal(raw, &gql); err != nil {
			return nil, fmt.Errorf("shopee api: resposta inválida: %w (corpo: %.200s)", err, string(raw))
		}
		if len(gql.Errors) > 0 {
			e := gql.Errors[0]
			return nil, fmt.Errorf("shopee api: erro %d: %s", e.Extensions.Code, e.Message)
		}

		for _, n := range gql.Data.ProductOfferV2.Nodes {
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

		if !gql.Data.ProductOfferV2.PageInfo.HasNextPage {
			break
		}
	}
	return produtos, nil
}
