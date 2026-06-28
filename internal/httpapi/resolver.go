package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// resolverLink segue redirects de um link curto da Shopee e extrai dados do produto.
// Se possível, busca dados completos (preço, comissão, imagem) via productOfferV2.
// POST /api/resolver-link  body: {"url": "https://s.shopee.com.br/HASH"}
func (srv *Server) resolverLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeErr(w, http.StatusBadRequest, "informe {\"url\": \"...\"}")
		return
	}

	// Segue redirects para obter a URL final
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 {
				return apperr.ErrTooManyRedirects
			}
			return nil
		},
	}

	resp, err := client.Head(req.URL)
	if err != nil {
		resp, err = client.Get(req.URL)
		if err != nil {
			srv.Logger.Error("resolver-link falhou", slog.String("url", req.URL), slog.String("erro", err.Error()))
			writeErr(w, http.StatusBadGateway, "não consegui resolver o link")
			return
		}
		defer resp.Body.Close()
	} else {
		defer resp.Body.Close()
	}

	urlFinal := resp.Request.URL.String()
	nome, shopID, itemID := extrairDadosURL(urlFinal)

	result := map[string]any{
		"url_final": urlFinal,
		"nome":      nome,
		"shop_id":   shopID,
		"item_id":   itemID,
	}

	// Se temos itemId, busca dados completos na API de afiliados (preço, imagem, etc.)
	if itemID != "" {
		appID := os.Getenv("SHOPEE_APP_ID")
		secret := os.Getenv("SHOPEE_SECRET")
		if appID != "" && secret != "" {
			if produto := buscarProdutoPorID(appID, secret, itemID); produto != nil {
				result["nome"] = produto.Name
				result["preco"] = produto.Price
				result["comissao"] = produto.Commission
				result["vendas"] = produto.Sales30d
				result["nota"] = produto.Rating
				result["imagem"] = produto.Image
				result["link_afiliado"] = produto.Link
			}
		}
	}

	srv.Logger.Info("resolver-link",
		slog.String("original", req.URL),
		slog.String("final", urlFinal),
		slog.String("item_id", itemID),
	)

	writeJSON(w, http.StatusOK, result)
}

var reShopeeProduct = regexp.MustCompile(`/([^/]+)-i\.(\d+)\.(\d+)`)

// extrairDadosURL extrai nome, shopID e itemID de uma URL longa da Shopee.
// Ex.: /Sérum-Vitamina-C-30ml-i.123456.789012 → ("Sérum Vitamina C 30ml", "123456", "789012")
func extrairDadosURL(url string) (nome, shopID, itemID string) {
	match := reShopeeProduct.FindStringSubmatch(url)
	if len(match) < 4 {
		// Tenta formato alternativo sem -i.
		// shopee.com.br/product/SHOP_ID/ITEM_ID
		re2 := regexp.MustCompile(`/product/(\d+)/(\d+)`)
		m2 := re2.FindStringSubmatch(url)
		if len(m2) >= 3 {
			return "", m2[1], m2[2]
		}
		return "", "", ""
	}

	nome = strings.ReplaceAll(match[1], "-", " ")
	// Remove possíveis query params do nome
	if idx := strings.Index(nome, "?"); idx > 0 {
		nome = nome[:idx]
	}
	return nome, match[2], match[3]
}

// buscarProdutoPorID consulta productOfferV2 com itemId para obter dados completos.
func buscarProdutoPorID(appID, secret, itemID string) *domain.Product {
	src := source.NewShopeeAPISource(appID, secret)
	src.ItemID = itemID
	src.Limit = 1
	src.MaxPages = 1

	produtos, err := src.Fetch()
	if err != nil || len(produtos) == 0 {
		return nil
	}
	return &produtos[0]
}
