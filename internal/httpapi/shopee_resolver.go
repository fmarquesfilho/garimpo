package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ── Regex para parsing de inputs de loja Shopee ───────────────────────────

// reShopIDURL casa com https://shopee.com.br/shop/123456
var reShopIDURL = regexp.MustCompile(`^https?://shopee\.com\.br/shop/(\d+)`)

// reSlugURL casa com https://shopee.com.br/{slug}
var reSlugURL = regexp.MustCompile(`^https?://shopee\.com\.br/([a-zA-Z0-9._-]+)`)

// reNumericID casa com IDs numéricos puros (5-15 dígitos)
var reNumericID = regexp.MustCompile(`^\d{5,15}$`)

// reShortLink casa com links curtos da Shopee (s.shopee.com.br/HASH)
var reShortLink = regexp.MustCompile(`^https?://s\.shopee\.com\.br/`)

// reProductURL casa com URLs de produto que contêm shop_id: /Nome-i.SHOP_ID.ITEM_ID
var reProductURL = regexp.MustCompile(`-i\.(\d+)\.\d+`)

// pathsReservados são segmentos de URL Shopee que não são slugs de loja
var pathsReservados = map[string]bool{
	"shop": true, "product": true, "m": true, "daily_discover": true,
	"search": true, "cart": true, "checkout": true, "user": true,
}

// parseShopInputWithName extrai o shop ID e tenta obter o nome da loja.
func (srv *Server) parseShopInputWithName(input string) (int64, string, error) {
	// Limpa query params e fragments
	input = cleanURL(input)

	// 0. Short link (s.shopee.com.br/HASH) — resolve redirect primeiro
	if reShortLink.MatchString(input) {
		resolved, err := srv.resolveShortLink(input)
		if err != nil {
			return 0, "", fmt.Errorf("não consegui resolver o link curto: %w", err)
		}
		input = cleanURL(resolved)
	}

	// 1. URL com /shop/{id}
	if m := reShopIDURL.FindStringSubmatch(input); len(m) == 2 {
		id, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, "", err
		}
		nome := srv.buscarNomeLoja(id)
		return id, nome, nil
	}

	// 2. URL de produto com -i.SHOP_ID.ITEM_ID — extrai o shop_id
	if m := reProductURL.FindStringSubmatch(input); len(m) == 2 {
		id, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, "", err
		}
		nome := srv.buscarNomeLoja(id)
		return id, nome, nil
	}

	// 3. URL com slug (https://shopee.com.br/{slug})
	if m := reSlugURL.FindStringSubmatch(input); len(m) == 2 {
		slug := m[1]
		if pathsReservados[slug] {
			return 0, "", fmt.Errorf("'%s' é um caminho reservado da Shopee, não um slug de loja", slug)
		}
		id, nome, err := srv.resolveShopSlugWithName(input)
		if err != nil {
			return 0, "", fmt.Errorf("não consegui encontrar o ID da loja '%s'. Tente copiar a URL no formato shopee.com.br/shop/123456 ou use o ID numérico", slug)
		}
		return id, nome, nil
	}

	// 4. ID numérico puro
	if reNumericID.MatchString(input) {
		id, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return 0, "", err
		}
		nome := srv.buscarNomeLoja(id)
		return id, nome, nil
	}

	return 0, "", fmt.Errorf("formato não reconhecido. Aceitos: URL da Shopee (shopee.com.br/shop/ID, link curto s.shopee.com.br/..., ou link de produto) ou ID numérico (5-15 dígitos)")
}

// buscarNomeLoja consulta a API pública da Shopee para obter o nome da loja via shopId.
func (srv *Server) buscarNomeLoja(shopID int64) string {
	client := &http.Client{Timeout: 5 * time.Second}
	apiURL := fmt.Sprintf("https://shopee.com.br/api/v4/shop/get_shop_detail?shopid=%d", shopID)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	return result.Data.Name
}

// resolveShortLink segue redirects de um link curto da Shopee e retorna a URL final.
func (srv *Server) resolveShortLink(shortURL string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 {
				return fmt.Errorf("muitos redirects")
			}
			return nil
		},
	}

	resp, err := client.Head(shortURL)
	if err != nil {
		// Fallback: tenta GET se HEAD falhar
		resp, err = client.Get(shortURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
	} else {
		defer resp.Body.Close()
	}

	return resp.Request.URL.String(), nil
}

// resolveShopSlugWithName resolve slug → (shopId, nome) via API pública v4 da Shopee.
func (srv *Server) resolveShopSlugWithName(shopURL string) (int64, string, error) {
	m := reSlugURL.FindStringSubmatch(cleanURL(shopURL))
	if len(m) < 2 {
		return 0, "", fmt.Errorf("slug não encontrado na URL")
	}
	slug := m[1]

	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := "https://shopee.com.br/api/v4/shop/get_shop_detail?username=" + slug

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("falha ao consultar Shopee: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Error    int    `json:"error"`
		ErrorMsg string `json:"error_msg"`
		Data     struct {
			ShopID int64  `json:"shopid"`
			Name   string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", fmt.Errorf("resposta inválida da Shopee: %w", err)
	}
	if result.Error != 0 {
		return 0, "", fmt.Errorf("Shopee retornou erro: %s", result.ErrorMsg)
	}
	if result.Data.ShopID <= 0 {
		return 0, "", fmt.Errorf("shopId não encontrado para '%s'", slug)
	}

	return result.Data.ShopID, result.Data.Name, nil
}

// cleanURL remove query params e fragments de uma URL.
func cleanURL(input string) string {
	// Remove query params
	if idx := strings.IndexByte(input, '?'); idx >= 0 {
		input = input[:idx]
	}
	// Remove fragments
	if idx := strings.IndexByte(input, '#'); idx >= 0 {
		input = input[:idx]
	}
	// Strip trailing slashes
	input = strings.TrimRight(input, "/")
	return input
}
