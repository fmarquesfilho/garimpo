package httpapi

import (
"context"
"encoding/json"
"log/slog"
"net/http"
"strconv"
"time"

"github.com/fmarquesfilho/garimpo/internal/engine"
"github.com/fmarquesfilho/garimpo/internal/scheduler"
"github.com/fmarquesfilho/garimpo/internal/store"
)

func (srv *Server) coletar(w http.ResponseWriter, r *http.Request) {
	if !srv.autorizadoColeta(r) {
		writeErr(w, http.StatusUnauthorized, "token de coleta inválido")
		return
	}

	q := r.URL.Query()
	estrategia := "nicho"
	if v := q.Get("estrategia"); v != "" {
		estrategia = v
	}

	src, chave := srv.fonte(q)
	produtos, err := srv.fetchCacheado(src, chave)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	scored := engine.Rankear(produtos, strategyDe(estrategia), srv.elegibilidade(q))
	n := topN(q)
	if n > len(scored) {
		n = len(scored)
	}

	categoria := q.Get("categoria")
	if categoria == "" {
		categoria = srv.Categoria
	}
	keyword := q.Get("keyword")
	if keyword == "" {
		keyword = srv.Keyword
	}

	snap := store.Snapshot{
		Categoria:  categoria,
		Keyword:    keyword,
		Estrategia: estrategia,
		Em:         time.Now().UTC(),
	}
	for i, s := range scored[:n] {
		p := s.Product
		snap.Itens = append(snap.Itens, store.ItemSnapshot{
Posicao: i + 1, ProdutoID: p.ID, Nome: p.Name,
Preco: p.Price, Comissao: p.Commission, Vendas: p.Sales30d,
Nota: p.Rating, Score: s.Score,
})
	}

	if err := srv.Eventos.RegistrarSnapshot(r.Context(), snap); err != nil {
		srv.Logger.Error("coleta falhou ao gravar snapshot",
slog.String("categoria", categoria), slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	srv.Logger.Info("coleta",
slog.String("categoria", categoria),
slog.String("keyword", keyword),
slog.String("estrategia", estrategia),
slog.Int("coletados", len(snap.Itens)),
slog.String("store", srv.Eventos.Nome()),
	)

	writeJSON(w, http.StatusAccepted, map[string]any{
"categoria":  categoria,
"estrategia": estrategia,
"coletados":  len(snap.Itens),
"em":         snap.Em,
})
}

func (srv *Server) listarBuscas(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeJSON(w, http.StatusOK, map[string]any{"buscas": []store.Busca{}, "store": srv.Eventos.Nome()})
		return
	}
	lista, err := srv.Eventos.ListarBuscas(r.Context())
	if err != nil {
		srv.Logger.Error("listar buscas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	var filtrada []store.Busca
	for _, b := range lista {
		if b.OwnerUID == "" || b.OwnerUID == user.UID {
			filtrada = append(filtrada, b)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"buscas": filtrada, "store": srv.Eventos.Nome()})
}

func (srv *Server) salvarBusca(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para salvar buscas")
		return
	}
	var b store.Busca
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	b = store.NormalizarBusca(b)
	if b.ID == "" {
		writeErr(w, http.StatusBadRequest, "busca precisa de ao menos uma keyword")
		return
	}
	b.Ativo = !r.URL.Query().Has("remover")
	b.OwnerUID = user.UID

	if err := srv.Eventos.SalvarBusca(r.Context(), b); err != nil {
		srv.Logger.Error("salvar busca falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	go func() {
		params := scheduler.ColetaParams{
			Categoria: b.Categoria, Estrategia: b.Estrategia,
			Top: b.Top, VendasMin: b.VendasMin, NotaMin: b.NotaMin, ShopIDs: b.ShopIDs,
		}
		var err error
		if b.Ativo {
			err = srv.Scheduler.SyncBusca(context.Background(), b.ID, b.Keywords, b.Cron, params)
		} else {
			err = srv.Scheduler.DeletarBusca(context.Background(), b.ID, b.Keywords)
		}
		if err != nil {
			srv.Logger.Error("scheduler sync falhou", slog.String("busca", b.ID), slog.String("erro", err.Error()))
		} else {
			srv.Logger.Info("scheduler sync", slog.String("busca", b.ID), slog.Bool("ativo", b.Ativo), slog.String("cron", b.Cron))
		}
	}()

	srv.Logger.Info("busca salva", slog.String("id", b.ID), slog.Bool("ativo", b.Ativo))
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "ok", "id": b.ID, "ativo": b.Ativo})
}

func (srv *Server) estatisticas(w http.ResponseWriter, r *http.Request) {
	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}
	est, err := srv.Eventos.Estatisticas(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("estatisticas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, est)
}

func (srv *Server) coletas(w http.ResponseWriter, r *http.Request) {
	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}
	historico, err := srv.Eventos.HistoricoColetas(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("historico coletas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"coletas": historico, "dias": dias})
}
