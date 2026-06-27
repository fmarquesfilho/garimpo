package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
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
	origemCacheTTL = 24 * time.Hour
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
	if len(origemCache) > 5000 {
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
	Fonte  string `json:"fonte"`
	Erro   string `json:"erro,omitempty"`
}

// produtoOrigem busca origem e marca de um produto.
// GET /api/produto/origem?item_id=123&shop_id=456
func (srv *Server) produtoOrigem(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("item_id")
	shopID := r.URL.Query().Get("shop_id")
	if itemID == "" || shopID == "" {
		writeErr(w, http.StatusBadRequest, "item_id e shop_id são obrigatórios")
		return
	}

	cacheKey := shopID + ":" + itemID

	if cached, ok := origemDoCache(cacheKey); ok {
		writeJSON(w, http.StatusOK, origemResponse{
			ItemID: itemID, ShopID: shopID,
			Origem: cached.Origem, Marca: cached.Marca, Fonte: "cache",
		})
		return
	}

	origem, marca, err := buscarOrigemProdutoShopee(itemID, shopID)
	if err != nil {
		srv.Logger.Warn("buscar origem falhou",
			slog.String("item_id", itemID),
			slog.String("shop_id", shopID),
			slog.String("erro", err.Error()),
		)
		writeJSON(w, http.StatusOK, origemResponse{
			ItemID: itemID, ShopID: shopID,
			Origem: "", Marca: "", Fonte: "erro", Erro: err.Error(),
		})
		return
	}

	origemNorm := NormalizarOrigemProduto(origem)
	salvarOrigemNoCache(cacheKey, origemCacheEntry{Origem: origemNorm, Marca: marca})

	writeJSON(w, http.StatusOK, origemResponse{
		ItemID: itemID, ShopID: shopID,
		Origem: origemNorm, Marca: marca, Fonte: "proxy_residencial",
	})
}

// produtoOrigemBatch resolve origem de múltiplos produtos.
// POST /api/produto/origem/batch
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
		req.Itens = req.Itens[:20]
	}

	resultados := make([]origemResponse, 0, len(req.Itens))
	for _, item := range req.Itens {
		if item.ItemID == "" || item.ShopID == "" {
			continue
		}
		cacheKey := item.ShopID + ":" + item.ItemID
		if cached, ok := origemDoCache(cacheKey); ok {
			resultados = append(resultados, origemResponse{
				ItemID: item.ItemID, ShopID: item.ShopID,
				Origem: cached.Origem, Marca: cached.Marca, Fonte: "cache",
			})
			continue
		}

		origem, marca, err := buscarOrigemProdutoShopee(item.ItemID, item.ShopID)
		if err != nil {
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
			Origem: origemNorm, Marca: marca, Fonte: "proxy_residencial",
		})
		time.Sleep(200 * time.Millisecond)
	}

	writeJSON(w, http.StatusOK, map[string]any{"resultados": resultados})
}

// ── Busca via proxy residencial ──────────────────────────────────────────────

// buscarOrigemProdutoShopee busca a página HTML do produto na Shopee via proxy
// residencial e extrai País de Origem e Marca dos dados embutidos.
//
// Requer: RESIDENTIAL_PROXY_URL no ambiente.
// Formato: http://usuario:senha_country-br@geo.iproyal.com:12321
func buscarOrigemProdutoShopee(itemID, shopID string) (origem string, marca string, err error) {
	proxyURL := os.Getenv("RESIDENTIAL_PROXY_URL")
	if proxyURL == "" {
		return "", "", fmt.Errorf("RESIDENTIAL_PROXY_URL não configurado")
	}

	transport := &http.Transport{}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return "", "", fmt.Errorf("proxy URL inválida: %w", err)
	}
	transport.Proxy = http.ProxyURL(parsed)

	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	// Tenta a página HTML do produto (contém dados SSR)
	productURL := fmt.Sprintf("https://shopee.com.br/product-i.%s.%s", shopID, itemID)
	req, err := http.NewRequest(http.MethodGet, productURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

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

	origem, marca = extrairOrigemDeHTML(string(body))

	// Fallback: tenta API JSON com o mesmo proxy
	if origem == "" && marca == "" {
		origem, marca = tentarAPIJSON(client, itemID, shopID)
	}

	return origem, marca, nil
}

// extrairOrigemDeHTML extrai País de Origem e Marca do HTML da página de produto.
func extrairOrigemDeHTML(html string) (origem, marca string) {
	// 1. Busca em atributos JSON embutidos no HTML (formato {"name":"...","value":"..."})
	attrPattern := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"\s*,\s*"value"\s*:\s*"([^"]+)"`)
	allAttrs := attrPattern.FindAllStringSubmatch(html, -1)
	for _, m := range allAttrs {
		if len(m) < 3 {
			continue
		}
		name := strings.ToLower(m[1])
		value := m[2]
		if origem == "" && (strings.Contains(name, "origem") || strings.Contains(name, "origin") ||
			strings.Contains(name, "país") || strings.Contains(name, "envio de")) {
			origem = value
		}
		if marca == "" && (strings.Contains(name, "marca") || strings.Contains(name, "brand")) {
			marca = value
		}
	}

	// 2. JSON-LD (schema.org brand)
	if marca == "" {
		brandPattern := regexp.MustCompile(`"brand"\s*:\s*\{\s*"name"\s*:\s*"([^"]+)"`)
		if m := brandPattern.FindStringSubmatch(html); len(m) >= 2 {
			marca = m[1]
		}
	}
	if marca == "" {
		brandSimple := regexp.MustCompile(`"brand"\s*:\s*"([^"]+)"`)
		if m := brandSimple.FindStringSubmatch(html); len(m) >= 2 {
			marca = m[1]
		}
	}

	// 3. Regex na tabela de especificações renderizada
	if origem == "" {
		origemHTML := regexp.MustCompile(`(?i)Pa[ií]s\s+de\s+Origem[^<]*</[^>]*>\s*<[^>]*>([^<]+)`)
		if m := origemHTML.FindStringSubmatch(html); len(m) >= 2 {
			origem = strings.TrimSpace(m[1])
		}
	}

	return origem, marca
}

// tentarAPIJSON tenta o endpoint API v4 com o mesmo client (proxy residencial).
func tentarAPIJSON(client *http.Client, itemID, shopID string) (origem, marca string) {
	apiURL := fmt.Sprintf("https://shopee.com.br/api/v4/item/get?itemid=%s&shopid=%s", itemID, shopID)
	req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
	if req == nil {
		return "", ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://shopee.com.br/")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", ""
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			ItemBasic struct {
				Brand      string `json:"brand"`
				Attributes []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"attributes"`
			} `json:"item_basic"`
		} `json:"data"`
		Item struct {
			Brand    string `json:"brand"`
			Location string `json:"location"`
		} `json:"item"`
	}
	if json.Unmarshal(raw, &result) != nil {
		return "", ""
	}

	for _, attr := range result.Data.ItemBasic.Attributes {
		name := strings.ToLower(attr.Name)
		if origem == "" && (strings.Contains(name, "origem") || strings.Contains(name, "origin") || strings.Contains(name, "país")) {
			origem = attr.Value
		}
		if marca == "" && (strings.Contains(name, "marca") || strings.Contains(name, "brand")) {
			marca = attr.Value
		}
	}
	if marca == "" && result.Data.ItemBasic.Brand != "" {
		marca = result.Data.ItemBasic.Brand
	}
	if marca == "" && result.Item.Brand != "" {
		marca = result.Item.Brand
	}
	if origem == "" && result.Item.Location != "" {
		origem = result.Item.Location
	}

	return origem, marca
}

// NormalizarOrigemProduto normaliza o valor de origem para formato padrão PT-BR.
func NormalizarOrigemProduto(origem string) string {
	if origem == "" {
		return ""
	}
	loc := strings.TrimSpace(strings.ToLower(origem))

	mapa := map[string]string{
		"coreia": "Coreia", "coréia": "Coreia", "korea": "Coreia",
		"south korea": "Coreia", "coreia do sul": "Coreia", "coréia do sul": "Coreia", "kr": "Coreia",
		"república da coreia": "Coreia",
		"japão": "Japão", "japao": "Japão", "japan": "Japão", "jp": "Japão",
		"china": "China", "mainland china": "China", "cn": "China",
		"república popular da china": "China",
		"brasil": "Brasil", "brazil": "Brasil", "br": "Brasil",
		"eua": "EUA", "usa": "EUA", "united states": "EUA", "estados unidos": "EUA",
		"taiwan": "Taiwan", "tw": "Taiwan",
		"tailândia": "Tailândia", "thailand": "Tailândia",
		"indonésia": "Indonésia", "indonesia": "Indonésia",
		"frança": "França", "france": "França",
		"itália": "Itália", "italy": "Itália",
	}

	if nome, ok := mapa[loc]; ok {
		return nome
	}
	if len(loc) > 0 {
		return strings.ToUpper(loc[:1]) + loc[1:]
	}
	return ""
}
