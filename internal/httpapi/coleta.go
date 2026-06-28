package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/fmarquesfilho/garimpo/internal/coleta"
	"github.com/fmarquesfilho/garimpo/internal/scheduler"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

func (srv *Server) coletar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parse dos parâmetros HTTP → struct de domínio
	params := coleta.Params{
		Estrategia:  q.Get("estrategia"),
		Categoria:   q.Get("categoria"),
		Keyword:     q.Get("keyword"),
		Top:         topN(q),
		BuscaID:     q.Get("busca_id"),
		VendasMin:   srv.VendasMin,
		NotaMin:     srv.NotaMin,
		ComissaoMin: 0.07,
	}
	if params.Estrategia == "" {
		params.Estrategia = "nicho"
	}
	if params.Categoria == "" {
		params.Categoria = srv.Categoria
	}
	if params.Keyword == "" {
		params.Keyword = srv.Keyword
	}
	if s := q.Get("comissao_min"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			params.ComissaoMin = v
		}
	}
	if s := q.Get("vendas_min"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			params.VendasMin = v
		}
	}
	if s := q.Get("nota_min"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			params.NotaMin = v
		}
	}

	// Resolve a fonte de produtos
	src, _ := srv.fonte(q)

	// Delega toda a lógica ao Service
	svc := coleta.Novo(coleta.Deps{
		Repo:   srv.Repo,
		Logger: srv.Logger,
	})

	resultado, err := svc.Executar(r.Context(), src, params)
	if err != nil {
		srv.Logger.Error("coleta falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resultado)
}

func (srv *Server) listarBuscas(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)
	lista, err := srv.Repo.Buscas().ListarBuscas(r.Context())
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
	writeJSON(w, http.StatusOK, map[string]any{"buscas": filtrada, "store": srv.Repo.Nome()})
}

func (srv *Server) salvarBusca(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)
	var b store.Busca
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	b = store.NormalizarBusca(b)
	// Uma busca precisa de ao menos um critério útil
	if len(b.Keywords) == 0 && len(b.ShopIDs) == 0 && len(b.Categorias) == 0 && len(b.Fontes) == 0 {
		writeErr(w, http.StatusBadRequest, "busca precisa de ao menos uma keyword, loja, categoria ou fonte")
		return
	}
	b.Ativo = !r.URL.Query().Has("remover")
	b.OwnerUID = user.UID

	if err := srv.Repo.Buscas().SalvarBusca(r.Context(), b); err != nil {
		srv.Logger.Error("salvar busca falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	// Scheduler sync síncrono
	params := scheduler.ColetaParams{
		BuscaID: b.ID, Categoria: b.Categoria, Estrategia: b.Estrategia,
		Top: b.Top, VendasMin: b.VendasMin, NotaMin: b.NotaMin, ShopIDs: b.ShopIDs,
	}
	if b.Ativo {
		if err := srv.Scheduler.SyncBusca(r.Context(), b.ID, b.Keywords, b.Cron, params); err != nil {
			srv.Logger.Error("scheduler sync falhou", slog.String("busca", b.ID), slog.String("erro", err.Error()))
		} else {
			srv.Logger.Info("scheduler sync", slog.String("busca", b.ID), slog.Bool("ativo", b.Ativo), slog.String("cron", b.Cron))
		}
	} else {
		if err := srv.Scheduler.DeletarBusca(r.Context(), b.ID, b.Keywords); err != nil {
			srv.Logger.Error("scheduler delete falhou", slog.String("busca", b.ID), slog.String("erro", err.Error()))
		}
	}

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
	est, err := srv.Repo.Snapshots().Estatisticas(r.Context(), dias)
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
	historico, err := srv.Repo.Snapshots().HistoricoColetas(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("historico coletas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"coletas": historico, "dias": dias})
}
