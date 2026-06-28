package httpapi

import (
	"encoding/json"
	"fmt"
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
}

// produtoOrigem retorna a origem de um produto a partir do cache ou
// da configuração de origem_padrao da loja monitorada.
// GET /api/produto/origem?item_id=123&shop_id=456
func (srv *Server) produtoOrigem(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("item_id")
	shopID := r.URL.Query().Get("shop_id")
	if itemID == "" || shopID == "" {
		writeErr(w, http.StatusBadRequest, "item_id e shop_id são obrigatórios")
		return
	}

	cacheKey := shopID + ":" + itemID

	// 1. Cache hit
	if cached, ok := origemDoCache(cacheKey); ok {
		writeJSON(w, http.StatusOK, origemResponse{
			ItemID: itemID, ShopID: shopID,
			Origem: cached.Origem, Marca: cached.Marca, Fonte: "cache",
		})
		return
	}

	// 2. Buscar origem_padrao da loja monitorada que contenha este shopId
	origem := srv.buscarOrigemPadraoDoShop(r, shopID)

	if origem != "" {
		salvarOrigemNoCache(cacheKey, origemCacheEntry{Origem: origem})
	}

	writeJSON(w, http.StatusOK, origemResponse{
		ItemID: itemID, ShopID: shopID,
		Origem: origem, Fonte: "loja_monitorada",
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

		origem := srv.buscarOrigemPadraoDoShop(r, item.ShopID)
		if origem != "" {
			salvarOrigemNoCache(cacheKey, origemCacheEntry{Origem: origem})
		}
		resultados = append(resultados, origemResponse{
			ItemID: item.ItemID, ShopID: item.ShopID,
			Origem: origem, Fonte: "loja_monitorada",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"resultados": resultados})
}

// buscarOrigemPadraoDoShop procura nas buscas monitoradas se o shopId
// pertence a uma loja com origem_padrao configurado.
func (srv *Server) buscarOrigemPadraoDoShop(r *http.Request, shopID string) string {
	buscas, err := srv.Repo.Buscas().ListarBuscas(r.Context())
	if err != nil {
		return ""
	}
	for _, b := range buscas {
		if !b.Ativo || b.OrigemPadrao == "" {
			continue
		}
		for _, sid := range b.ShopIDs {
			if strings.TrimSpace(shopID) == strings.TrimSpace(fmt.Sprintf("%d", sid)) {
				return b.OrigemPadrao
			}
		}
	}
	return ""
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
		"japão": "Japão", "japao": "Japão", "japan": "Japão", "jp": "Japão",
		"china": "China", "mainland china": "China", "cn": "China",
		"brasil": "Brasil", "brazil": "Brasil", "br": "Brasil",
		"eua": "EUA", "usa": "EUA", "united states": "EUA",
		"taiwan":    "Taiwan",
		"tailândia": "Tailândia", "thailand": "Tailândia",
		"indonésia": "Indonésia", "indonesia": "Indonésia",
	}

	if nome, ok := mapa[loc]; ok {
		return nome
	}
	if len(loc) > 0 {
		return strings.ToUpper(loc[:1]) + loc[1:]
	}
	return ""
}
