package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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

	// Cache: novidades mudam a cada coleta (~4h), então 5 min de TTL é seguro
	cacheKey := buscaID + ":" + strconv.Itoa(dias)
	srv.muNov.Lock()
	if srv.cacheNov == nil {
		srv.cacheNov = make(map[string]*cacheEntryNov)
	}
	if e, ok := srv.cacheNov[cacheKey]; ok && time.Since(e.em) < 5*time.Minute {
		srv.muNov.Unlock()
		writeJSON(w, http.StatusOK, e.dados)
		return
	}
	srv.muNov.Unlock()

	novidades, err := srv.Eventos.Novidades(r.Context(), buscaID, dias)
	if err != nil {
		srv.Logger.Error("novidades falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	srv.muNov.Lock()
	srv.cacheNov[cacheKey] = &cacheEntryNov{dados: novidades, em: time.Now()}
	srv.muNov.Unlock()

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

	input := cleanURL(req.Input)
	if input == "" {
		writeErr(w, http.StatusBadRequest, "informe uma URL ou ID de loja")
		return
	}

	shopID, shopName, err := srv.parseShopInputWithName(req.Input)
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

	// Registra job no scheduler
	params := scheduler.ColetaParams{
		BuscaID:    busca.ID,
		Estrategia: busca.Estrategia,
		ShopIDs:    busca.ShopIDs,
		Top:        100,
	}
	if err := srv.Scheduler.SyncBusca(r.Context(), busca.ID, busca.Keywords, busca.Cron, params); err != nil {
		srv.Logger.Error("scheduler sync loja falhou",
			slog.String("busca", busca.ID), slog.String("erro", err.Error()))
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
