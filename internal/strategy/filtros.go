package strategy

import "github.com/fmarquesfilho/garimpo/internal/domain"

// Filtro é uma etapa do pipeline de elegibilidade. Cada filtro decide se um
// produto passa ou não, de forma independente.
// Padrão: Chain of Responsibility — cada filtro pode ser composto/encadeado.
type Filtro interface {
	// Aceita retorna true se o produto deve permanecer na lista.
	Aceita(p domain.Product) bool
	// Nome identifica o filtro (para logging/debug).
	Nome() string
}

// Pipeline é uma lista de filtros aplicados em sequência.
// Um produto precisa passar por TODOS os filtros para ser elegível.
// Com pipeline vazio, tudo passa (modo sem filtro).
type Pipeline []Filtro

// Aplicar filtra o slice de produtos, retornando apenas os aceitos por todos.
func (pl Pipeline) Aplicar(produtos []domain.Product) []domain.Product {
	if len(pl) == 0 {
		return produtos // sem filtros = tudo passa
	}
	out := make([]domain.Product, 0, len(produtos))
	for _, p := range produtos {
		if pl.aceita(p) {
			out = append(out, p)
		}
	}
	return out
}

func (pl Pipeline) aceita(p domain.Product) bool {
	for _, f := range pl {
		if !f.Aceita(p) {
			return false
		}
	}
	return true
}

// ─── Filtros concretos ────────────────────────────────────────────────────────

// FiltroComissao descarta produtos abaixo de um piso de comissão.
type FiltroComissao struct{ Min float64 }

func (f FiltroComissao) Nome() string                 { return "comissao" }
func (f FiltroComissao) Aceita(p domain.Product) bool { return p.Commission >= f.Min }

// FiltroVendas descarta produtos com vendas abaixo de um piso.
type FiltroVendas struct{ Min int }

func (f FiltroVendas) Nome() string                 { return "vendas" }
func (f FiltroVendas) Aceita(p domain.Product) bool { return f.Min <= 0 || p.Sales30d >= f.Min }

// FiltroNota descarta produtos com nota abaixo de um piso.
type FiltroNota struct{ Min float64 }

func (f FiltroNota) Nome() string                 { return "nota" }
func (f FiltroNota) Aceita(p domain.Product) bool { return f.Min <= 0 || p.Rating >= f.Min }

// FiltroPrecoMax descarta produtos acima de um teto de preço.
type FiltroPrecoMax struct{ Max float64 }

func (f FiltroPrecoMax) Nome() string                 { return "preco_max" }
func (f FiltroPrecoMax) Aceita(p domain.Product) bool { return f.Max <= 0 || p.Price <= f.Max }

// FiltroPrecoMin descarta produtos abaixo de um piso de preço.
type FiltroPrecoMin struct{ Min float64 }

func (f FiltroPrecoMin) Nome() string                 { return "preco_min" }
func (f FiltroPrecoMin) Aceita(p domain.Product) bool { return f.Min <= 0 || p.Price >= f.Min }

// ─── Builders ─────────────────────────────────────────────────────────────────

// PipelineCuradoria monta o pipeline padrão para curadoria (filtros da Elegibilidade).
func PipelineCuradoria(e Elegibilidade) Pipeline {
	var pl Pipeline
	if e.ComissaoMin > 0 {
		pl = append(pl, FiltroComissao{Min: e.ComissaoMin})
	}
	if e.VendasMin > 0 {
		pl = append(pl, FiltroVendas{Min: e.VendasMin})
	}
	if e.NotaMin > 0 {
		pl = append(pl, FiltroNota{Min: e.NotaMin})
	}
	return pl
}

// PipelineMonitoramento retorna um pipeline vazio (sem filtros) — mostra tudo
// que a API da Shopee retornou. Os filtros são aplicados opcionalmente pelo
// usuário no frontend.
func PipelineMonitoramento() Pipeline {
	return Pipeline{} // vazio = tudo passa
}
