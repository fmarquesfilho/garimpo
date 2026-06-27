package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/scheduler"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// novidades compara os últimos dois snapshots de uma busca com lojas e retorna:
// - produtos novos (presentes no último snapshot mas não no anterior)
// - variações de preço (mesmo produto_id com preço diferente)
func (srv *Server) novidades(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver novidades")
		return
	}

	buscaID := r.URL.Query().Get("busca_id")
	dias := 7
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}

	novidades, err := srv.Eventos.Novidades(r.Context(), buscaID, dias)
	if err != nil {
		srv.Logger.Error("novidades falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, novidades)
}

// evolucaoLojas retorna a evolução de preço das lojas monitoradas ao longo do tempo.
func (srv *Server) evolucaoLojas(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver evolução de lojas")
		return
	}

	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}

	evolucao, err := srv.Eventos.EvolucaoLojas(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("evolução lojas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, evolucao)
}

// ── Adicionar loja ────────────────────────────────────────────────────────

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

type adicionarLojaReq struct {
	Input        string `json:"input"`         // URL ou ID numérico
	Cron         string `json:"cron"`          // cron expression (opcional, default "0 */4 * * *")
	OrigemPadrao string `json:"origem_padrao"` // origem padrão dos produtos (ex: "Coreia", "Japão")
}

type adicionarLojaResp struct {
	Status string `json:"status"`
	ID     string `json:"id"`
	ShopID int64  `json:"shop_id"`
	Nome   string `json:"nome"`
}

// adicionarLoja aceita uma URL de loja Shopee ou ID numérico, extrai o shopID
// e cria uma Busca com shop_ids para monitoramento.
func (srv *Server) adicionarLoja(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para adicionar lojas")
		return
	}

	var req adicionarLojaReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	input := strings.TrimSpace(req.Input)
	if input == "" {
		writeErr(w, http.StatusBadRequest, "informe uma URL ou ID de loja")
		return
	}

	shopID, shopName, err := srv.parseShopInputWithName(input)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	// Verifica duplicata: busca ativa com mesmo shop_id
	buscas, _ := srv.Eventos.ListarBuscas(r.Context())
	for _, b := range buscas {
		if !b.Ativo || b.OwnerUID != user.UID {
			continue
		}
		for _, sid := range b.ShopIDs {
			if sid == shopID {
				writeErr(w, http.StatusConflict, fmt.Sprintf("loja %d já está sendo monitorada (busca: %s)", shopID, b.ID))
				return
			}
		}
	}

	cron := req.Cron
	if cron == "" {
		cron = "0 */4 * * *"
	}

	busca := store.NormalizarBusca(store.Busca{
		ShopIDs:      []int64{shopID},
		Nome:         shopName,
		Estrategia:   "nicho",
		Cron:         cron,
		Ativo:        true,
		OwnerUID:     user.UID,
		OrigemPadrao: req.OrigemPadrao,
	})

	if err := srv.Eventos.SalvarBusca(r.Context(), busca); err != nil {
		srv.Logger.Error("adicionar loja falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	// Registra job no scheduler (síncrono para garantir criação antes do Cloud Run matar a instância)
	params := scheduler.ColetaParams{
		BuscaID:    busca.ID,
		Estrategia: busca.Estrategia,
		ShopIDs:    busca.ShopIDs,
		Top:        100,
	}
	if err := srv.Scheduler.SyncBusca(r.Context(), busca.ID, busca.Keywords, busca.Cron, params); err != nil {
		srv.Logger.Error("scheduler sync loja falhou",
			slog.String("busca", busca.ID), slog.String("erro", err.Error()))
		// Não falha a request — a busca foi salva, o scheduler pode ser recriado depois
	}

	srv.Logger.Info("loja adicionada",
		slog.String("busca_id", busca.ID),
		slog.String("nome", shopName),
		slog.Int64("shop_id", shopID),
		slog.String("owner", user.UID),
	)

	writeJSON(w, http.StatusCreated, adicionarLojaResp{
		Status: "ok",
		ID:     busca.ID,
		ShopID: shopID,
		Nome:   shopName,
	})
}

// parseShopInputWithName extrai o shop ID e tenta obter o nome da loja.
func (srv *Server) parseShopInputWithName(input string) (int64, string, error) {
	// Limpa query params e fragments
	input = cleanURL(input)

	// 0. Short link (s.shopee.com.br/HASH) — resolve redirect primeiro
	if reShortLink.MatchString(input) {
		resolved, err := srv.resolveShortLink(input)
		if err != nil {
			return 0, "", fmt.Errorf("não consegui resolver o link curto: %v", err)
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
		return 0, "", fmt.Errorf("falha ao consultar Shopee: %v", err)
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
		return 0, "", fmt.Errorf("resposta inválida da Shopee: %v", err)
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

// listarLojas retorna as buscas ativas do usuário que têm shop_ids.
func (srv *Server) listarLojas(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver lojas")
		return
	}

	buscas, err := srv.Eventos.ListarBuscas(r.Context())
	if err != nil {
		srv.Logger.Error("listar lojas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	var lojas []store.Busca
	for _, b := range buscas {
		if b.Ativo && (b.OwnerUID == "" || b.OwnerUID == user.UID) && len(b.ShopIDs) > 0 {
			lojas = append(lojas, b)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"lojas": lojas})
}

// removerLoja desativa uma busca de loja (tombstone).
func (srv *Server) removerLoja(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para remover lojas")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "parâmetro 'id' é obrigatório")
		return
	}

	// Busca a busca existente para confirmar ownership
	buscas, err := srv.Eventos.ListarBuscas(r.Context())
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	var encontrada *store.Busca
	for _, b := range buscas {
		if b.ID == id && b.OwnerUID == user.UID {
			encontrada = &b
			break
		}
	}
	if encontrada == nil {
		writeErr(w, http.StatusNotFound, "loja não encontrada")
		return
	}

	encontrada.Ativo = false
	if err := srv.Eventos.SalvarBusca(r.Context(), *encontrada); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	// Remove do scheduler
	go func() {
		_ = srv.Scheduler.DeletarBusca(r.Context(), encontrada.ID, encontrada.Keywords)
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "removida", "id": id})
}
