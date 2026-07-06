package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// resolveShortLink segue redirects de links curtos (s.shopee.com.br/xxx) e retorna a URL final.
func resolveShortLink(ctx context.Context, shortURL string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Para no primeiro redirect para shopee.com.br (evita loops)
			if strings.Contains(req.URL.Host, "shopee.com.br") && !strings.Contains(req.URL.Host, "s.shopee.com.br") {
				return http.ErrUseLastResponse
			}
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, shortURL, nil)
	if err != nil {
		return "", fmt.Errorf("criando request para link curto: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("seguindo redirect do link curto: %w", err)
	}
	defer resp.Body.Close()

	// O redirect final está no Location header ou na URL do response
	if loc := resp.Header.Get("Location"); loc != "" {
		return loc, nil
	}
	return resp.Request.URL.String(), nil
}

func (s *UnifiedCollectorServer) ResolveShop(ctx context.Context, req *collectorpb.ResolveShopRequest) (*collectorpb.ResolveShopResponse, error) {
	if req.GetUsernameOrUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "username_or_url é obrigatório")
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	marketplace := source.ProtoToMarketplace(mkt)

	if marketplace != domain.MarketplaceShopee {
		return nil, status.Errorf(codes.Unimplemented, "ResolveShop não suportado para %q", marketplace)
	}

	username := req.GetUsernameOrUrl()
	if strings.HasPrefix(username, "http") {
		u, err := url.Parse(username)
		if err == nil {
			// Links curtos (s.shopee.com.br/xxx) precisam de redirect para resolver
			if strings.Contains(u.Host, "s.shopee.com.br") {
				resolved, err := resolveShortLink(ctx, username)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "falha ao resolver link curto: %v", err)
				}
				u, _ = url.Parse(resolved)
			}
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) > 0 {
				username = parts[len(parts)-1]
			}
		}
	}

	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "username inválido")
	}

	apiURL := "https://shopee.com.br/api/v4/shop/get_shop_detail?username=" + url.QueryEscape(username)
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao criar request: %v", err)
	}
	reqHTTP.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(reqHTTP)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro na requisição Shopee API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao ler resposta: %v", err)
	}

	var data struct {
		Error int `json:"error"`
		Data  struct {
			ShopID int64  `json:"shopid"`
			Name   string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao fazer parse do JSON: %v", err)
	}

	if data.Error != 0 || data.Data.ShopID == 0 {
		return nil, status.Errorf(codes.NotFound, "loja não encontrada (error=%d)", data.Error)
	}

	return &collectorpb.ResolveShopResponse{
		ShopId:   data.Data.ShopID,
		ShopName: data.Data.Name,
	}, nil
}

func (s *UnifiedCollectorServer) GenerateAffiliateLink(ctx context.Context, req *collectorpb.GenerateAffiliateLinkRequest) (*collectorpb.GenerateAffiliateLinkResponse, error) {
	if req.GetOriginalUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "original_url é obrigatório")
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	marketplace := source.ProtoToMarketplace(mkt)

	if marketplace != domain.MarketplaceShopee {
		return nil, status.Errorf(codes.Unimplemented, "GenerateAffiliateLink não suportado para %q", marketplace)
	}

	// Busca credenciais do primeiro receiver Shopee configurado
	appID, secret := s.pipeline.ShopeeCredentials()
	if appID == "" || secret == "" {
		return nil, status.Error(codes.FailedPrecondition, "credenciais Shopee não configuradas")
	}

	// Monta query GraphQL generateShortLink
	subIDsJSON := "[]"
	if len(req.GetSubIds()) > 0 {
		parts := make([]string, 0, len(req.GetSubIds()))
		for _, sid := range req.GetSubIds() {
			parts = append(parts, `"`+sid+`"`)
		}
		subIDsJSON = "[" + strings.Join(parts, ",") + "]"
	}

	query := fmt.Sprintf(
		`{ generateShortLink(originUrl: "%s", subIds: %s) { shortLink } }`,
		req.GetOriginalUrl(), subIDsJSON,
	)

	payload, _ := json.Marshal(map[string]string{"query": query})

	ts := fmt.Sprintf("%d", time.Now().Unix())
	sum := sha256.Sum256([]byte(appID + ts + string(payload) + secret))
	sig := hex.EncodeToString(sum[:])

	endpoint := "https://open-api.affiliate.shopee.com.br/graphql"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao criar request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", appID, ts, sig))

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro na requisição Shopee API: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao ler resposta: %v", err)
	}

	var gql struct {
		Data struct {
			GenerateShortLink struct {
				ShortLink string `json:"shortLink"`
			} `json:"generateShortLink"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &gql); err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao fazer parse do JSON: %v", err)
	}

	if len(gql.Errors) > 0 {
		return nil, status.Errorf(codes.Internal, "Shopee API error: %s (code=%d)", gql.Errors[0].Message, gql.Errors[0].Code)
	}

	return &collectorpb.GenerateAffiliateLinkResponse{
		ShortLink: gql.Data.GenerateShortLink.ShortLink,
		LongLink:  req.GetOriginalUrl(),
	}, nil
}
