package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ── Cache de origem ──────────────────────────────────────────────────────────

type origemCacheEntry struct {
	Origem string
	Marca  string
	Em     time.Time
}

var (
	origemCache    = make(map[string]origemCacheEntry)
	origemCacheMu  sync.RWMutex
	origemCacheTTL = 24 * time.Hour // cache por 24h (origem não muda)
)

func origemDoCache(key string) (origemCacheEntry, bool) {
	origemCacheMu.RLock()
	defer origemCacheMu.RUnlock()
	e, ok := origemCache[key]
	if !ok || time.Since(e.Em) > origemCacheTTL {
		return origemCacheEntry{}, false
	}
	return e, true
}

func salvarOrigemNoCache(key string, e origemCacheEntry) {
	origemCacheMu.Lock()
	defer origemCacheMu.Unlock()
	e.Em = time.Now()
	origemCache[key] = e
	// Limita o cache a 5000 entradas (evitar memory leak)
	if len(origemCache) > 5000 {
		// Remove entradas mais antigas (simples: limpa tudo e recomeça)
		origemCache = make(map[string]origemCacheEntry)
		origemCache[key] = e
	}
}

// ── Handler ──────────────────────────────────────────────────────────────────

type origemResponse struct {
	ItemID string `json:"item_id"`
	ShopID string `json:"shop_id"`
	Origem string `json:"origem"`
	Marca  string `json:"marca,omitempty"`
	Fonte  string `json:"fonte"`          // "api_publica" | "cache" | "fallback" | "erro"
	Erro   string `json:"erro,omitempty"` // detalhe do erro (debug)
}

// produtoOrigem consulta a API pública v4 da Shopee para obter os atributos
// do produto (incluindo País de Origem e Marca).
// GET /api/produto/origem?item_id=123&shop_id=456
func (srv *Server) produtoOrigem(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("item_id")
	shopID := r.URL.Query().Get("shop_id")
	if itemID == "" || shopID == "" {
		writeErr(w, http.StatusBadRequest, "item_id e shop_id são obrigatórios")
		return
	}

	cacheKey := shopID + ":" + itemID

	// 1. Verificar cache
	if cached, ok := origemDoCache(cacheKey); ok {
		writeJSON(w, http.StatusOK, origemResponse{
			ItemID: itemID, ShopID: shopID,
			Origem: cached.Origem, Marca: cached.Marca, Fonte: "cache",
		})
		return
	}

	// 2. Chamar API pública v4 da Shopee
	origem, marca, err := buscarOrigemProdutoShopee(itemID, shopID)
	if err != nil {
		srv.Logger.Warn("buscar origem falhou",
			slog.String("item_id", itemID),
			slog.String("shop_id", shopID),
			slog.String("erro", err.Error()),
		)
		// Retorna vazio mas não falha (graceful degradation)
		writeJSON(w, http.StatusOK, origemResponse{
			ItemID: itemID, ShopID: shopID,
			Origem: "", Marca: "", Fonte: "erro",
			Erro: err.Error(),
		})
		return
	}

	// 3. Normalizar e cachear
	origemNorm := NormalizarOrigemProduto(origem)
	salvarOrigemNoCache(cacheKey, origemCacheEntry{Origem: origemNorm, Marca: marca})

	writeJSON(w, http.StatusOK, origemResponse{
		ItemID: itemID, ShopID: shopID,
		Origem: origemNorm, Marca: marca, Fonte: "api_publica",
	})
}

// produtoOrigemBatch resolve origem de múltiplos produtos de uma vez.
// POST /api/produto/origem/batch  body: {"itens": [{"item_id": "123", "shop_id": "456"}, ...]}
func (srv *Server) produtoOrigemBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Itens []struct {
			ItemID string `json:"item_id"`
			ShopID string `json:"shop_id"`
		} `json:"itens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if len(req.Itens) == 0 {
		writeErr(w, http.StatusBadRequest, "lista de itens vazia")
		return
	}
	if len(req.Itens) > 20 {
		req.Itens = req.Itens[:20] // Limita a 20 por request
	}

	resultados := make([]origemResponse, 0, len(req.Itens))

	for _, item := range req.Itens {
		if item.ItemID == "" || item.ShopID == "" {
			continue
		}
		cacheKey := item.ShopID + ":" + item.ItemID

		// Cache hit?
		if cached, ok := origemDoCache(cacheKey); ok {
			resultados = append(resultados, origemResponse{
				ItemID: item.ItemID, ShopID: item.ShopID,
				Origem: cached.Origem, Marca: cached.Marca, Fonte: "cache",
			})
			continue
		}

		// Buscar na API (com throttle leve entre chamadas)
		origem, marca, err := buscarOrigemProdutoShopee(item.ItemID, item.ShopID)
		if err != nil {
			srv.Logger.Warn("batch origem falhou",
				slog.String("item_id", item.ItemID),
				slog.String("erro", err.Error()),
			)
			resultados = append(resultados, origemResponse{
				ItemID: item.ItemID, ShopID: item.ShopID,
				Origem: "", Marca: "", Fonte: "erro",
			})
			continue
		}

		origemNorm := NormalizarOrigemProduto(origem)
		salvarOrigemNoCache(cacheKey, origemCacheEntry{Origem: origemNorm, Marca: marca})

		resultados = append(resultados, origemResponse{
			ItemID: item.ItemID, ShopID: item.ShopID,
			Origem: origemNorm, Marca: marca, Fonte: "api_publica",
		})

		// Throttle: 150ms entre chamadas à API pública
		time.Sleep(150 * time.Millisecond)
	}

	writeJSON(w, http.StatusOK, map[string]any{"resultados": resultados})
}

// ── Chamada à API pública v4 ─────────────────────────────────────────────────

// buscarOrigemProdutoShopee consulta a API pública v4 da Shopee Brasil para
// obter "País de Origem" e "Marca" dos atributos do produto.
// Endpoint: https://shopee.com.br/api/v4/pdp/get_pc?item_id=X&shop_id=Y
func buscarOrigemProdutoShopee(itemID, shopID string) (origem string, marca string, err error) {
	url := fmt.Sprintf("https://shopee.com.br/api/v4/pdp/get_pc?item_id=%s&shop_id=%s", itemID, shopID)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	// Headers necessários para a API pública aceitar a requisição
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://shopee.com.br/")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("request falhou: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Parse da resposta — a estrutura é aninhada
	var result struct {
		Data struct {
			Product struct {
				Brand string `json:"brand"`
				// Atributos ficam em tier_variations ou em item_basic.attributes
			} `json:"product"`
			ItemBasic struct {
				Brand      string `json:"brand"`
				BrandID    int64  `json:"brand_id"`
				Attributes []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"attributes"`
			} `json:"item_basic"`
			// Formato alternativo: item.attributes
			Item struct {
				Brand      string `json:"brand"`
				Attributes []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"attributes"`
				Location string `json:"location"`
			} `json:"item"`
		} `json:"data"`
		// Formato alternativo (v4/item/get)
		Item struct {
			Brand    string `json:"brand"`
			Location string `json:"location"`
			Props    []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"props"`
			SellerInfo struct {
				City string `json:"city"`
			} `json:"seller_info"`
		} `json:"item"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("parse falhou: %w", err)
	}

	// Extrair origem de múltiplas fontes possíveis (a estrutura varia)
	// 1. Atributos do produto (item_basic.attributes ou item.attributes)
	attrs := result.Data.ItemBasic.Attributes
	if len(attrs) == 0 {
		attrs = result.Data.Item.Attributes
	}
	if len(attrs) == 0 && result.Item.Props != nil {
		for _, p := range result.Item.Props {
			attrs = append(attrs, struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			}{Name: p.Name, Value: p.Value})
		}
	}

	for _, attr := range attrs {
		name := strings.ToLower(attr.Name)
		// Campos que indicam origem
		if strings.Contains(name, "origem") || strings.Contains(name, "origin") ||
			strings.Contains(name, "país") || strings.Contains(name, "country") ||
			strings.Contains(name, "envio de") || strings.Contains(name, "fabricado") {
			if attr.Value != "" && origem == "" {
				origem = attr.Value
			}
		}
		// Campos que indicam marca
		if strings.Contains(name, "marca") || strings.Contains(name, "brand") {
			if attr.Value != "" && marca == "" {
				marca = attr.Value
			}
		}
	}

	// 2. Campo "brand" direto
	if marca == "" {
		if result.Data.ItemBasic.Brand != "" {
			marca = result.Data.ItemBasic.Brand
		} else if result.Data.Product.Brand != "" {
			marca = result.Data.Product.Brand
		} else if result.Data.Item.Brand != "" {
			marca = result.Data.Item.Brand
		} else if result.Item.Brand != "" {
			marca = result.Item.Brand
		}
	}

	// 3. Campo "location" (localidade do vendedor/produto)
	if origem == "" {
		if result.Data.Item.Location != "" {
			origem = result.Data.Item.Location
		} else if result.Item.Location != "" {
			origem = result.Item.Location
		} else if result.Item.SellerInfo.City != "" {
			origem = result.Item.SellerInfo.City
		}
	}

	return origem, marca, nil
}

// NormalizarOrigemProduto normaliza o valor de origem para formato padrão PT-BR.
func NormalizarOrigemProduto(origem string) string {
	if origem == "" {
		return ""
	}
	loc := strings.TrimSpace(strings.ToLower(origem))

	mapa := map[string]string{
		// Coreano
		"coreia":              "Coreia",
		"coréia":              "Coreia",
		"korea":               "Coreia",
		"south korea":         "Coreia",
		"coreia do sul":       "Coreia",
		"coréia do sul":       "Coreia",
		"kr":                  "Coreia",
		"república da coreia": "Coreia",
		// Japonês
		"japão": "Japão",
		"japao": "Japão",
		"japan": "Japão",
		"jp":    "Japão",
		// Chinês
		"china":                      "China",
		"mainland china":             "China",
		"cn":                         "China",
		"república popular da china": "China",
		// Brasileiro
		"brasil": "Brasil",
		"brazil": "Brasil",
		"br":     "Brasil",
		// Outros
		"eua":            "EUA",
		"usa":            "EUA",
		"united states":  "EUA",
		"estados unidos": "EUA",
		"taiwan":         "Taiwan",
		"tw":             "Taiwan",
		"tailândia":      "Tailândia",
		"thailand":       "Tailândia",
		"indonésia":      "Indonésia",
		"indonesia":      "Indonésia",
		"frança":         "França",
		"france":         "França",
		"itália":         "Itália",
		"italy":          "Itália",
	}

	if nome, ok := mapa[loc]; ok {
		return nome
	}

	// Capitaliza a primeira letra se não encontrou no mapa
	if len(loc) > 0 {
		return strings.ToUpper(loc[:1]) + loc[1:]
	}
	return ""
}
