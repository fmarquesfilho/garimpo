// Package store registra eventos (decisões de curadoria, e depois conversões)
// para análise posterior. A interface isola o destino: NopStore em dev/local,
// BigQueryStore em produção (atrás da build tag `gcp`, para não pesar o build
// padrão nem o CI com a dependência do BigQuery).
package store

import (
	"context"
	"time"
)

// Evento é um fato registrado. As tags JSON casam com o objeto `candidato` que
// o front já manipula, então dá para postar o candidato direto + um `tipo`.
type Evento struct {
	Tipo       string    `json:"tipo"` // ex.: "selecao", "publicacao"
	ProdutoID  string    `json:"id"`
	Nome       string    `json:"nome"`
	Categoria  string    `json:"categoria"`
	Estrategia string    `json:"estrategia"`
	Canal      string    `json:"canal,omitempty"` // preenchido em publicações
	Comissao   float64   `json:"comissao"`
	Preco      float64   `json:"preco"`
	Vendas     int       `json:"vendas"`
	Score      float64   `json:"score"`
	Em         time.Time `json:"-"` // carimbado no servidor
}

// EventoStore registra eventos. Implementações: NopStore e BigQueryStore.
type EventoStore interface {
	Registrar(ctx context.Context, e Evento) error
	RegistrarSnapshot(ctx context.Context, s Snapshot) error
	Nome() string
}

// ItemSnapshot é um produto na foto de mercado de uma categoria, num instante.
type ItemSnapshot struct {
	Posicao   int
	ProdutoID string
	Nome      string
	Preco     float64
	Comissao  float64
	Vendas    int
	Nota      float64
	Score     float64
}

// Snapshot é a foto periódica de uma categoria: os top N do momento. É o que
// permite, depois, analisar o impacto das campanhas contra o pano de fundo do
// mercado (preço/comissão/demanda mudaram em volta da publicação?).
type Snapshot struct {
	Categoria  string
	Keyword    string
	Estrategia string
	Em         time.Time
	Itens      []ItemSnapshot
}

// NopStore descarta eventos — usado localmente e quando o BigQuery não está ligado.
type NopStore struct{}

func (NopStore) Registrar(context.Context, Evento) error           { return nil }
func (NopStore) RegistrarSnapshot(context.Context, Snapshot) error { return nil }
func (NopStore) Nome() string                                      { return "nop" }
