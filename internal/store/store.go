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
	Canal      string    `json:"canal,omitempty"`  // preenchido em publicações
	SubID      string    `json:"sub_id,omitempty"` // atribuição (canal_estrategia_data)
	Comissao   float64   `json:"comissao"`
	Preco      float64   `json:"preco"`
	Vendas     int       `json:"vendas"`
	Score      float64   `json:"score"`
	Em         time.Time `json:"-"` // carimbado no servidor
}

// EstatCategoria resume os snapshots de mercado de uma categoria na janela.
// É a primeira camada descritiva do pipeline de análise (Fase 1 da estratégia):
// "como está cada categoria agora", sobre o dado que já se acumula no tempo.
type EstatCategoria struct {
	Categoria       string  `json:"categoria"`
	Amostras        int     `json:"amostras"`
	ComissaoMedia   float64 `json:"comissao_media"`
	ComissaoMediana float64 `json:"comissao_mediana"`
	PrecoMedio      float64 `json:"preco_medio"`
	VendasMedia     float64 `json:"vendas_media"`
	TeorMedio       float64 `json:"teor_medio"`
}

// Estatisticas é o resumo descritivo dos snapshots coletados numa janela.
type Estatisticas struct {
	Fonte         string           `json:"fonte"`          // "bigquery" | "nop"
	DiasJanela    int              `json:"dias_janela"`    // janela considerada
	TotalAmostras int              `json:"total_amostras"` // itens de snapshot
	PorCategoria  []EstatCategoria `json:"por_categoria"`  // agregado por categoria
	GeradoEm      time.Time        `json:"gerado_em"`
}

// Busca é um "perfil de coleta": um conjunto nomeado de filtros que pode ser
// reusado manualmente (no front) e rodado periodicamente (Cloud Scheduler) para
// coletar snapshots. Cada busca carrega seu próprio `Cron` (periodicidade), o
// que torna a coleta agendada configurável por perfil e independente do navegador.
type Busca struct {
	Nome        string    `json:"nome"`
	Keyword     string    `json:"keyword"`
	Categoria   string    `json:"categoria"`
	Estrategia  string    `json:"estrategia"`
	ComissaoMin float64   `json:"comissao_min"`
	VendasMin   int       `json:"vendas_min"`
	NotaMin     float64   `json:"nota_min"`
	Top         int       `json:"top"`
	Cron        string    `json:"cron"`  // ex.: "0 8 * * *" (vazio = só manual)
	Ativo       bool      `json:"ativo"` // false = removida (tombstone)
	SalvoEm     time.Time `json:"salvo_em"`
}

// EventoStore registra eventos. Implementações: NopStore e BigQueryStore.
type EventoStore interface {
	Registrar(ctx context.Context, e Evento) error
	RegistrarSnapshot(ctx context.Context, s Snapshot) error
	// Estatisticas devolve o resumo descritivo dos snapshots dos últimos `dias`.
	Estatisticas(ctx context.Context, dias int) (Estatisticas, error)
	// SalvarBusca persiste (append) um perfil de busca; ListarBuscas devolve o
	// estado atual (último registro por nome, só os ativos).
	SalvarBusca(ctx context.Context, b Busca) error
	ListarBuscas(ctx context.Context) ([]Busca, error)
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
func (NopStore) Estatisticas(_ context.Context, dias int) (Estatisticas, error) {
	// Sem persistência local: devolve um resumo vazio, deixando claro a fonte.
	return Estatisticas{Fonte: "nop", DiasJanela: dias, GeradoEm: time.Now().UTC()}, nil
}

// SalvarBusca/ListarBuscas no Nop são no-op: localmente, as buscas vivem no
// navegador (localStorage). O sync server-side só acontece com o BigQuery ligado.
func (NopStore) SalvarBusca(context.Context, Busca) error      { return nil }
func (NopStore) ListarBuscas(context.Context) ([]Busca, error) { return nil, nil }
func (NopStore) Nome() string                                  { return "nop" }
