// Package coleta encapsula a lógica de execução de uma coleta:
// buscar produtos → aplicar rotação → rankear → gravar snapshot → alertar.
// Separado do HTTP handler para ser testável independentemente.
package coleta

import (
	"context"
	"log/slog"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/alerts"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/engine"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
	"github.com/fmarquesfilho/garimpo/internal/strategy"
)

// Params define os parâmetros de uma coleta.
type Params struct {
	Estrategia  string
	Categoria   string
	Keyword     string
	Top         int
	BuscaID     string // se preenchido, habilita rotação e alertas
	VendasMin   int
	NotaMin     float64
	ComissaoMin float64
}

// Resultado é o retorno de uma coleta executada com sucesso.
type Resultado struct {
	Categoria  string    `json:"categoria"`
	Estrategia string    `json:"estrategia"`
	Keyword    string    `json:"keyword"`
	Coletados  int       `json:"coletados"`
	Em         time.Time `json:"em"`
}

// Deps agrupa as dependências do service (injeção explícita).
type Deps struct {
	Store  store.EventoStore
	Logger *slog.Logger
}

// Service orquestra o fluxo completo de coleta.
type Service struct {
	deps Deps
}

// Novo cria um Service de coleta.
func Novo(deps Deps) *Service {
	return &Service{deps: deps}
}

// Executar roda uma coleta completa: fetch → rotação → rank → snapshot → alertas.
func (s *Service) Executar(ctx context.Context, src source.ProductSource, params Params) (Resultado, error) {
	// 1. Resolver busca (para rotação)
	var busca *store.Busca
	if params.BuscaID != "" {
		if shopSrc, ok := src.(*source.ShopeeShopSource); ok {
			busca = s.carregarBusca(ctx, params.BuscaID)
			if busca != nil {
				s.aplicarRotacao(shopSrc, busca)
			}
		}
	}

	// 2. Fetch de produtos
	produtos, err := src.Fetch()
	if err != nil {
		return Resultado{}, err
	}

	// 2.5 Fallback de origem: se a Busca tem OrigemPadrao, aplicar nos produtos sem Origin
	if busca != nil && busca.OrigemPadrao != "" {
		for i := range produtos {
			if produtos[i].Origin == "" {
				produtos[i].Origin = busca.OrigemPadrao
			}
		}
	}

	// 3. Atualizar cursor de rotação
	if busca != nil {
		s.atualizarCursor(ctx, src, busca)
	}

	// 4. Rankear (sempre usa estratégia nicho — diversificada foi descontinuada da UI)
	scored := engine.Rankear(produtos, strategy.NewNiche(), elegibilidade(params))
	n := params.Top
	if n <= 0 {
		n = 20
	}
	if n > len(scored) {
		n = len(scored)
	}

	// 5. Montar e gravar snapshot
	keyword := params.Keyword
	if keyword == "" && params.BuscaID != "" {
		keyword = params.BuscaID
	}

	snap := store.Snapshot{
		Categoria:  params.Categoria,
		Keyword:    keyword,
		Estrategia: params.Estrategia,
		Em:         time.Now().UTC(),
	}
	for i, sc := range scored[:n] {
		p := sc.Product
		snap.Itens = append(snap.Itens, store.ItemSnapshot{
			Posicao: i + 1, ProdutoID: p.ID, Nome: p.Name,
			Preco: p.Price, Comissao: p.Commission, Vendas: p.Sales30d,
			Nota: p.Rating, Score: sc.Score, Origin: p.Origin,
		})
	}

	if err := s.deps.Store.RegistrarSnapshot(ctx, snap); err != nil {
		return Resultado{}, err
	}

	s.deps.Logger.Info("coleta",
		slog.String("categoria", params.Categoria),
		slog.String("keyword", keyword),
		slog.String("estrategia", params.Estrategia),
		slog.Int("coletados", len(snap.Itens)),
	)

	// 6. Alertas (background — não bloqueia resposta)
	if params.BuscaID != "" {
		go s.dispararAlertas(params.BuscaID)
	}

	return Resultado{
		Categoria:  params.Categoria,
		Estrategia: params.Estrategia,
		Keyword:    keyword,
		Coletados:  len(snap.Itens),
		Em:         snap.Em,
	}, nil
}

// ── Métodos internos ─────────────────────────────────────────────────────────

func (s *Service) carregarBusca(ctx context.Context, buscaID string) *store.Busca {
	buscas, err := s.deps.Store.ListarBuscas(ctx)
	if err != nil {
		return nil
	}
	for _, b := range buscas {
		if b.ID == buscaID && b.Ativo {
			return &b
		}
	}
	return nil
}

func (s *Service) aplicarRotacao(shopSrc *source.ShopeeShopSource, busca *store.Busca) {
	if busca.RotationCursor != nil && len(busca.ShopIDs) > 0 {
		firstShop := busca.ShopIDs[0]
		if pg, exists := busca.RotationCursor[firstShop]; exists && pg > 1 {
			shopSrc.StartPage = pg
		}
	}
	// Throttling
	shopSrc.PageDelay = 200 * time.Millisecond
	shopSrc.ShopDelay = 60 * time.Second
}

func (s *Service) atualizarCursor(ctx context.Context, src source.ProductSource, busca *store.Busca) {
	shopSrc, ok := src.(*source.ShopeeShopSource)
	if !ok || shopSrc.LastPageInfo == nil {
		return
	}
	if busca.RotationCursor == nil {
		busca.RotationCursor = make(map[int64]int)
	}
	if busca.FullScanAt == nil {
		busca.FullScanAt = make(map[int64]string)
	}
	for shopID, info := range shopSrc.LastPageInfo {
		busca.RotationCursor[shopID] = info.NextPage
		if !info.HasMore {
			busca.FullScanAt[shopID] = time.Now().UTC().Format(time.RFC3339)
		}
	}
	// Persiste o cursor (best-effort, não bloqueia)
	go func() {
		if err := s.deps.Store.SalvarBusca(context.Background(), *busca); err != nil {
			s.deps.Logger.Error("atualizar rotation cursor falhou",
				slog.String("busca", busca.ID), slog.String("erro", err.Error()))
		}
	}()
}

func (s *Service) dispararAlertas(buscaID string) {
	cfg := alerts.ConfigFromEnv()
	cfg.Logger = s.deps.Logger
	if !cfg.Ativo() {
		return
	}
	alerter := alerts.Novo(cfg)
	alerter.VerificarENotificar(context.Background(), s.deps.Store, buscaID)
	alerter.VerificarNovos(context.Background(), s.deps.Store, buscaID)
}

// ── Helpers (extraídos do httpapi para reuso) ─────────────────────────────────

func elegibilidade(p Params) strategy.Elegibilidade {
	e := strategy.Elegibilidade{
		ComissaoMin: 0.07,
		VendasMin:   p.VendasMin,
		NotaMin:     p.NotaMin,
	}
	if p.ComissaoMin > 0 {
		e.ComissaoMin = p.ComissaoMin
	}
	return e
}

// FetchDireto busca produtos sem cache (para uso fora do HTTP handler).
func FetchDireto(src source.ProductSource) ([]domain.Product, error) {
	return src.Fetch()
}
