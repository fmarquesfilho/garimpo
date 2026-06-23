package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// resolverLink segue redirects de um link curto da Shopee e extrai dados do produto.
// POST /api/resolver-link  body: {"url": "https://s.shopee.com.br/HASH"}
// Retorna: {url_final, nome, shop_id, item_id}
func (srv *Server) resolverLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeErr(w, http.StatusBadRequest, "informe {\"url\": \"...\"}")
		return
	}

	// Segue redirects para obter a URL final (sem baixar o corpo)
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 {
				return fmt.Errorf("muitos redirects")
			}
			return nil
		},
	}

	resp, err := client.Head(req.URL)
	if err != nil {
		// Alguns links curtos não respondem a HEAD, tenta GET
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

	// Extrai dados da URL final da Shopee
	// Formato: shopee.com.br/Nome-do-Produto-i.SHOP_ID.ITEM_ID
	nome, shopID, itemID := extrairDadosURL(urlFinal)

	result := map[string]any{
		"url_final": urlFinal,
		"nome":      nome,
		"shop_id":   shopID,
		"item_id":   itemID,
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
